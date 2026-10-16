package gormtx

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

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

func newDeferredFKManager(t *testing.T) *TxManager {
	t.Helper()

	dsn := fmt.Sprintf(
		"file:%s?_pragma=foreign_keys(1)",
		filepath.Join(
			t.TempDir(),
			"fk.db",
		),
	)

	db, err := gorm.Open(
		sqlite.Open(dsn),
		&gorm.Config{},
	)
	require.NoError(
		t,
		err,
	)

	sqlDB, err := db.DB()
	require.NoError(
		t,
		err,
	)

	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(
		func() {
			require.NoError(
				t,
				sqlDB.Close(),
			)
		},
	)

	require.NoError(
		t,
		db.AutoMigrate(&item{}),
	)
	require.NoError(
		t,
		db.Exec("CREATE TABLE parents (id INTEGER PRIMARY KEY)").Error,
	)
	require.NoError(
		t,
		db.Exec("CREATE TABLE children (id INTEGER PRIMARY KEY, parent_id INTEGER REFERENCES parents(id) DEFERRABLE INITIALLY DEFERRED)").Error,
	)

	manager, err := New(Opts{
		DB: db,
	})
	require.NoError(
		t,
		err,
	)

	return manager
}

func TestAfterCommitOutsideTransaction(t *testing.T) {
	manager := newTestManager(t)
	called := false
	effect := func(context.Context) error {
		called = true

		return nil
	}

	require.False(
		t,
		txcore.AddEffect(
			context.Background(),
			effect,
		),
	)

	err := manager.Do(
		context.Background(),
		"root",
		func(ctx context.Context) error {
			require.False(
				t,
				txcore.AddEffect(
					txcore.Detach(ctx),
					effect,
				),
			)

			return nil
		},
	)
	require.NoError(
		t,
		err,
	)
	require.False(
		t,
		called,
	)
}

func TestAfterCommitRunsAfterCommitInOrder(t *testing.T) {
	manager := newTestManager(t)
	log := &effectLog{}

	var seen []string

	inspect := func(ctx context.Context) error {
		require.False(
			t,
			manager.InTransaction(ctx),
		)
		require.False(
			t,
			txcore.AddEffect(
				ctx,
				log.effect("late"),
			),
		)

		seen = names(
			t,
			manager,
		)

		return log.effect("b")(ctx)
	}

	err := manager.Do(
		context.Background(),
		"root",
		func(ctx context.Context) error {
			require.True(
				t,
				txcore.AddEffect(
					ctx,
					log.effect("a"),
				),
			)
			require.True(
				t,
				txcore.AddEffect(
					ctx,
					inspect,
				),
			)
			require.True(
				t,
				txcore.AddEffect(
					ctx,
					log.effect("c"),
				),
			)
			require.Empty(
				t,
				log.names(),
			)

			return insert(
				ctx,
				manager,
				"row",
			)
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
		[]string{"row"},
		seen,
	)
}

func TestAfterCommitDiscardedWhenRootFails(t *testing.T) {
	failure := errors.New("root failed")

	tests := []struct {
		name    string
		manager func(t *testing.T) *TxManager
		fn      func(manager *TxManager) func(ctx context.Context) error
		wantErr error
		fnOK    bool
	}{
		{
			name:    "fn error",
			manager: newTestManager,
			fn: func(manager *TxManager) func(ctx context.Context) error {
				return func(context.Context) error {
					return failure
				}
			},
			wantErr: failure,
		},
		{
			name:    "fn panic",
			manager: newTestManager,
			fn: func(manager *TxManager) func(ctx context.Context) error {
				return func(context.Context) error {
					panic("boom")
				}
			},
			wantErr: ErrPanic,
		},
		{
			name:    "commit failure",
			manager: newDeferredFKManager,
			fn: func(manager *TxManager) func(ctx context.Context) error {
				return func(ctx context.Context) error {
					return manager.GetConnection(ctx).Exec("INSERT INTO children (parent_id) VALUES (42)").Error
				}
			},
			fnOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				manager := tt.manager(t)
				log := &effectLog{}
				body := tt.fn(manager)
				fnReturned := false

				var fnErr error

				err := manager.Do(
					context.Background(),
					"root",
					func(ctx context.Context) error {
						require.True(
							t,
							txcore.AddEffect(
								ctx,
								log.effect("effect"),
							),
						)

						fnErr = body(ctx)
						fnReturned = true

						return fnErr
					},
				)

				require.Error(
					t,
					err,
				)

				if tt.fnOK {
					require.True(
						t,
						fnReturned,
					)
					require.NoError(
						t,
						fnErr,
					)
				}

				if tt.wantErr != nil {
					require.ErrorIs(
						t,
						err,
						tt.wantErr,
					)
				}

				require.Empty(
					t,
					log.names(),
				)
			},
		)
	}
}

func TestAfterCommitOnlySuccessfulRetryAttempt(t *testing.T) {
	manager := newTestManager(t)
	log := &effectLog{}
	calls := 0

	err := manager.Do(
		context.Background(),
		"retry",
		func(ctx context.Context) error {
			calls++

			require.True(
				t,
				txcore.AddEffect(
					ctx,
					log.effect(fmt.Sprintf(
						"attempt-%d",
						calls,
					)),
				),
			)

			if calls == 1 {
				return &sqlStateError{
					code: "40001",
				}
			}

			return nil
		},
		txcore.WithRetry(2),
		txcore.WithBackoff(noBackoff),
	)

	require.NoError(
		t,
		err,
	)
	require.Equal(
		t,
		2,
		calls,
	)
	require.Equal(
		t,
		[]string{"attempt-2"},
		log.names(),
	)
}

func TestAfterCommitSavepointRollbackTruncatesEffects(t *testing.T) {
	manager := newTestManager(t)
	log := &effectLog{}
	failure := errors.New("nested failed")

	err := manager.Do(
		context.Background(),
		"root",
		func(ctx context.Context) error {
			require.True(
				t,
				txcore.AddEffect(
					ctx,
					log.effect("a"),
				),
			)

			nestedErr := manager.Do(
				ctx,
				"failing",
				func(ctx context.Context) error {
					require.True(
						t,
						txcore.AddEffect(
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

			require.NoError(
				t,
				manager.Do(
					ctx,
					"passing",
					func(ctx context.Context) error {
						require.True(
							t,
							txcore.AddEffect(
								ctx,
								log.effect("c"),
							),
						)

						l3Err := manager.Do(
							ctx,
							"deep",
							func(ctx context.Context) error {
								require.True(
									t,
									txcore.AddEffect(
										ctx,
										log.effect("d"),
									),
								)

								panic("boom")
							},
						)
						require.ErrorIs(
							t,
							l3Err,
							ErrPanic,
						)

						return nil
					},
				),
			)

			outerErr := manager.Do(
				ctx,
				"outer-failing",
				func(ctx context.Context) error {
					require.True(
						t,
						txcore.AddEffect(
							ctx,
							log.effect("e"),
						),
					)

					return manager.Do(
						ctx,
						"inner-passing",
						func(ctx context.Context) error {
							require.True(
								t,
								txcore.AddEffect(
									ctx,
									log.effect("f"),
								),
							)

							return failure
						},
					)
				},
			)
			require.ErrorIs(
				t,
				outerErr,
				failure,
			)

			require.True(
				t,
				txcore.AddEffect(
					ctx,
					log.effect("g"),
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
		[]string{"a", "c", "g"},
		log.names(),
	)
}

func TestAfterCommitEffectFailuresAreLogged(t *testing.T) {
	logger := &recordingLogger{}
	manager := newFileManager(
		t,
		Opts{
			Logger: logger,
		},
	)
	log := &effectLog{}

	err := manager.Do(
		context.Background(),
		"root",
		func(ctx context.Context) error {
			require.True(
				t,
				txcore.AddEffect(
					ctx,
					func(context.Context) error { return errors.New("effect failed") },
				),
			)
			require.True(
				t,
				txcore.AddEffect(
					ctx,
					func(context.Context) error { panic("effect boom") },
				),
			)
			require.True(
				t,
				txcore.AddEffect(
					ctx,
					log.effect("after"),
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
		[]string{"after"},
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
		logger.errorMsg,
	)
}

func TestAfterCommitRejectedAfterRootReturns(t *testing.T) {
	manager := newTestManager(t)
	log := &effectLog{}

	var escaped context.Context

	require.NoError(
		t,
		manager.Do(
			context.Background(),
			"root",
			func(ctx context.Context) error {
				escaped = ctx

				return nil
			},
		),
	)

	require.False(
		t,
		txcore.AddEffect(
			escaped,
			log.effect("late"),
		),
	)
	require.Empty(
		t,
		log.names(),
	)
}

func TestAfterCommitConcurrentRegistration(t *testing.T) {
	manager := newTestManager(t)
	log := &effectLog{}

	const workers = 32

	err := manager.Do(
		context.Background(),
		"root",
		func(ctx context.Context) error {
			var wg sync.WaitGroup

			for worker := range workers {
				wg.Add(1)

				go func() {
					defer wg.Done()

					txcore.AddEffect(
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
