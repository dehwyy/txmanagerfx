package txmanagertest

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager"
	"github.com/stretchr/testify/require"
)

type recordingLogger struct {
	mu     sync.Mutex
	warns  []string
	errors []string
}

func (l *recordingLogger) Warn(msg string, _ ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.warns = append(
		l.warns,
		msg,
	)
}

func (l *recordingLogger) Error(msg string, _ ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.errors = append(
		l.errors,
		msg,
	)
}

type effectLog struct {
	mu   sync.Mutex
	runs []string
}

func (l *effectLog) effect(name string) func(context.Context) error {
	return func(context.Context) error {
		l.mu.Lock()
		defer l.mu.Unlock()

		l.runs = append(
			l.runs,
			name,
		)

		return nil
	}
}

func (l *effectLog) names() []string {
	l.mu.Lock()
	defer l.mu.Unlock()

	return append(
		[]string(nil),
		l.runs...,
	)
}

func TestFakeAfterCommitRunsAfterCommitInOrder(t *testing.T) {
	logger := &recordingLogger{}
	fake := &Fake{
		Logger: logger,
	}
	log := &effectLog{}

	require.False(
		t,
		txmanager.AfterCommit(
			context.Background(),
			log.effect("outside"),
		),
	)

	inspect := func(ctx context.Context) error {
		require.False(
			t,
			fake.InTransaction(ctx),
		)
		require.False(
			t,
			txmanager.AfterCommit(
				ctx,
				log.effect("late"),
			),
		)

		return log.effect("b")(ctx)
	}

	err := fake.Do(
		context.Background(),
		"root",
		func(ctx context.Context) error {
			require.True(
				t,
				txmanager.AfterCommit(
					ctx,
					log.effect("a"),
				),
			)
			require.True(
				t,
				txmanager.AfterCommit(
					ctx,
					inspect,
				),
			)
			require.True(
				t,
				txmanager.AfterCommit(
					ctx,
					func(context.Context) error { return errors.New("effect failed") },
				),
			)
			require.True(
				t,
				txmanager.AfterCommit(
					ctx,
					func(context.Context) error { panic("effect boom") },
				),
			)
			require.True(
				t,
				txmanager.AfterCommit(
					ctx,
					log.effect("c"),
				),
			)
			require.Empty(
				t,
				log.names(),
			)

			return nil
		},
	)

	require.NoError(
		t,
		err,
	)
	require.Equal(
		t,
		[]string{"a", "b", "c"},
		log.names(),
	)
	require.Equal(
		t,
		[]string{"after-commit effect failed"},
		logger.warns,
	)
	require.Equal(
		t,
		[]string{"after-commit effect panicked"},
		logger.errors,
	)
}

func TestFakeAfterCommitDiscarded(t *testing.T) {
	failure := errors.New("failed")

	tests := []struct {
		name    string
		fake    *Fake
		fn      func(ctx context.Context, fake *Fake) error
		wantErr error
	}{
		{
			name:    "fn error",
			fake:    New(),
			fn:      func(context.Context, *Fake) error { return failure },
			wantErr: failure,
		},
		{
			name:    "fn panic",
			fake:    New(),
			fn:      func(context.Context, *Fake) error { panic("boom") },
			wantErr: txmanager.ErrPanic,
		},
		{
			name: "commit error",
			fake: &Fake{
				CommitErr: failure,
			},
			fn:      func(context.Context, *Fake) error { return nil },
			wantErr: failure,
		},
		{
			name: "rollback-only",
			fake: New(),
			fn: func(ctx context.Context, fake *Fake) error {
				nestedErr := fake.Do(
					txmanager.Flat(ctx),
					"nested",
					func(context.Context) error { return failure },
				)
				if !errors.Is(
					nestedErr,
					failure,
				) {
					return errors.New("unexpected nested result")
				}

				return nil
			},
			wantErr: txmanager.ErrFlatNestedFailed,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				log := &effectLog{}

				err := tt.fake.Do(
					context.Background(),
					"root",
					func(ctx context.Context) error {
						require.True(
							t,
							txmanager.AfterCommit(
								ctx,
								log.effect("effect"),
							),
						)

						return tt.fn(
							ctx,
							tt.fake,
						)
					},
				)

				require.ErrorIs(
					t,
					err,
					tt.wantErr,
				)
				require.Empty(
					t,
					log.names(),
				)
			},
		)
	}
}

func TestFakeFlatSwallowedFailure(t *testing.T) {
	logger := &recordingLogger{}
	fake := &Fake{
		Logger: logger,
	}
	failure := errors.New("nested failed")

	err := fake.Do(
		txmanager.Flat(context.Background()),
		"root",
		func(ctx context.Context) error {
			require.ErrorIs(
				t,
				fake.Do(
					ctx,
					"nested",
					func(context.Context) error { return failure },
				),
				failure,
			)
			require.ErrorIs(
				t,
				fake.Do(
					ctx,
					"panicking",
					func(context.Context) error { panic("boom") },
				),
				txmanager.ErrPanic,
			)

			return nil
		},
	)

	require.ErrorIs(
		t,
		err,
		txmanager.ErrFlatNestedFailed,
	)
	require.ErrorIs(
		t,
		err,
		failure,
	)
	require.Equal(
		t,
		[]string{"flat nested transaction failed, root rolled back"},
		logger.warns,
	)
	require.Equal(
		t,
		[]Call{
			{
				Name: "root",
				Flat: true,
			},
			{
				Name:   "nested",
				Nested: true,
				Flat:   true,
			},
			{
				Name:   "panicking",
				Nested: true,
				Flat:   true,
			},
		},
		fake.Calls(),
	)
}

func TestFakeFlatErrorReturnedByRootHasNoWarn(t *testing.T) {
	logger := &recordingLogger{}
	fake := &Fake{
		Logger: logger,
	}
	failure := errors.New("nested failed")

	err := fake.Do(
		txmanager.Flat(context.Background()),
		"root",
		func(ctx context.Context) error {
			return fake.Do(
				ctx,
				"nested",
				func(context.Context) error { return failure },
			)
		},
	)

	require.ErrorIs(
		t,
		err,
		failure,
	)
	require.NotErrorIs(
		t,
		err,
		txmanager.ErrFlatNestedFailed,
	)
	require.Empty(
		t,
		logger.warns,
	)
}

func TestFakeWithSavepointUnderFlat(t *testing.T) {
	logger := &recordingLogger{}
	fake := &Fake{
		Logger: logger,
	}
	log := &effectLog{}
	failure := errors.New("isolated failed")

	err := fake.Do(
		txmanager.Flat(context.Background()),
		"root",
		func(ctx context.Context) error {
			require.True(
				t,
				txmanager.AfterCommit(
					ctx,
					log.effect("a"),
				),
			)

			isolatedErr := fake.Do(
				ctx,
				"isolated",
				func(ctx context.Context) error {
					require.True(
						t,
						txmanager.AfterCommit(
							ctx,
							log.effect("isolated"),
						),
					)

					return failure
				},
				txmanager.WithSavepoint(),
			)
			require.ErrorIs(
				t,
				isolatedErr,
				failure,
			)

			absorbedErr := fake.Do(
				ctx,
				"absorbing",
				func(ctx context.Context) error {
					require.True(
						t,
						txmanager.AfterCommit(
							ctx,
							log.effect("absorbing"),
						),
					)
					require.ErrorIs(
						t,
						fake.Do(
							ctx,
							"flat",
							func(context.Context) error { return failure },
						),
						failure,
					)

					return nil
				},
				txmanager.WithSavepoint(),
			)
			require.ErrorIs(
				t,
				absorbedErr,
				txmanager.ErrFlatNestedFailed,
			)

			require.True(
				t,
				txmanager.AfterCommit(
					ctx,
					log.effect("b"),
				),
			)

			return nil
		},
	)

	require.NoError(
		t,
		err,
	)
	require.Equal(
		t,
		[]string{"a", "b"},
		log.names(),
	)
	require.Equal(
		t,
		[]string{"flat nested transaction failed, savepoint rolled back"},
		logger.warns,
	)
	require.Equal(
		t,
		[]Call{
			{
				Name: "root",
				Flat: true,
			},
			{
				Name:      "isolated",
				Nested:    true,
				Flat:      true,
				Savepoint: true,
			},
			{
				Name:      "absorbing",
				Nested:    true,
				Flat:      true,
				Savepoint: true,
			},
			{
				Name:   "flat",
				Nested: true,
				Flat:   true,
			},
		},
		fake.Calls(),
	)
}

func TestFakeSavepointRollbackTruncatesEffects(t *testing.T) {
	fake := New()
	log := &effectLog{}
	failure := errors.New("nested failed")

	err := fake.Do(
		context.Background(),
		"root",
		func(ctx context.Context) error {
			require.True(
				t,
				txmanager.AfterCommit(
					ctx,
					log.effect("a"),
				),
			)

			nestedErr := fake.Do(
				ctx,
				"nested",
				func(ctx context.Context) error {
					require.True(
						t,
						txmanager.AfterCommit(
							ctx,
							log.effect("b"),
						),
					)

					return failure
				},
			)
			require.ErrorIs(
				t,
				nestedErr,
				failure,
			)

			require.True(
				t,
				txmanager.AfterCommit(
					ctx,
					log.effect("c"),
				),
			)

			return nil
		},
	)

	require.NoError(
		t,
		err,
	)
	require.Equal(
		t,
		[]string{"a", "c"},
		log.names(),
	)
}

func TestFakeSubtransactionCounter(t *testing.T) {
	tests := []struct {
		name      string
		nested    int
		wantWarns []string
	}{
		{
			name:   "64",
			nested: 64,
		},
		{
			name:      "65",
			nested:    65,
			wantWarns: []string{"subtransaction count exceeded 64"},
		},
		{
			name:      "130",
			nested:    130,
			wantWarns: []string{"subtransaction count exceeded 64"},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				logger := &recordingLogger{}
				fake := &Fake{
					Logger: logger,
				}

				err := fake.Do(
					context.Background(),
					"root",
					func(ctx context.Context) error {
						for i := range tt.nested {
							if err := fake.Do(
								ctx,
								fmt.Sprintf(
									"n%d",
									i,
								),
								func(context.Context) error { return nil },
							); err != nil {
								return err
							}
						}

						return nil
					},
				)

				require.NoError(
					t,
					err,
				)
				require.Equal(
					t,
					tt.wantWarns,
					logger.warns,
				)
			},
		)
	}
}

func TestFakeDetachInsideFlatOpensOwnRoot(t *testing.T) {
	fake := New()
	log := &effectLog{}
	failure := errors.New("outer failed")

	err := fake.Do(
		txmanager.Flat(context.Background()),
		"outer",
		func(ctx context.Context) error {
			require.NoError(
				t,
				fake.Do(
					txmanager.Detach(ctx),
					"detached",
					func(ctx context.Context) error {
						require.True(
							t,
							txmanager.AfterCommit(
								ctx,
								log.effect("detached"),
							),
						)

						return nil
					},
				),
			)

			require.True(
				t,
				txmanager.AfterCommit(
					ctx,
					log.effect("outer"),
				),
			)

			return failure
		},
	)

	require.ErrorIs(
		t,
		err,
		failure,
	)
	require.Equal(
		t,
		[]string{"detached"},
		log.names(),
	)
	require.Equal(
		t,
		[]Call{
			{
				Name: "outer",
				Flat: true,
			},
			{
				Name: "detached",
			},
		},
		fake.Calls(),
	)
}

func TestFakeAfterCommitConcurrentRegistration(t *testing.T) {
	fake := New()
	log := &effectLog{}

	const workers = 32

	err := fake.Do(
		context.Background(),
		"root",
		func(ctx context.Context) error {
			var wg sync.WaitGroup

			for worker := range workers {
				wg.Add(1)

				go func() {
					defer wg.Done()

					txmanager.AfterCommit(
						ctx,
						log.effect(fmt.Sprintf(
							"w%d",
							worker,
						)),
					)
				}()
			}

			wg.Wait()

			return nil
		},
	)

	require.NoError(
		t,
		err,
	)
	require.Len(
		t,
		log.names(),
		workers,
	)
}
