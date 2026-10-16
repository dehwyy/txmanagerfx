package gormtx

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
	"github.com/stretchr/testify/require"
)

func TestFlatNestedIssuesNoSavepoint(t *testing.T) {
	tests := []struct {
		name      string
		beforeDo  bool
		insideDo  bool
		wantSaves int
	}{
		{
			name:      "marker before root",
			beforeDo:  true,
			wantSaves: 0,
		},
		{
			name:      "marker inside root",
			insideDo:  true,
			wantSaves: 0,
		},
		{
			name:      "no marker",
			wantSaves: 4,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				manager := newTestManager(t)
				log := recordSQL(
					t,
					manager,
				)
				ctx := context.Background()

				if tt.beforeDo {
					ctx = txcore.Flat(ctx)
				}

				require.False(
					t,
					manager.InTransaction(ctx),
				)

				err := manager.Do(
					ctx,
					"root",
					func(ctx context.Context) error {
						if tt.insideDo {
							ctx = txcore.Flat(ctx)
						}

						if err := insert(
							ctx,
							manager,
							"root",
						); err != nil {
							return err
						}

						for _, name := range []string{"m1", "m2"} {
							if err := manager.Do(
								ctx,
								name,
								func(ctx context.Context) error {
									return insert(
										ctx,
										manager,
										name,
									)
								},
							); err != nil {
								return err
							}
						}

						return manager.Do(
							ctx,
							"outer",
							func(ctx context.Context) error {
								return manager.Do(
									ctx,
									"inner",
									func(ctx context.Context) error {
										return insert(
											ctx,
											manager,
											"inner",
										)
									},
								)
							},
						)
					},
				)

				require.NoError(
					t,
					err,
				)
				require.Equal(
					t,
					tt.wantSaves,
					log.count(sqlSavepoint),
				)
				require.Zero(
					t,
					log.count(sqlRollbackTo),
				)
				assertNames(
					t,
					manager,
					"root",
					"m1",
					"m2",
					"inner",
				)
			},
		)
	}
}

func TestFlatNestedErrorReturnedByRoot(t *testing.T) {
	logger := &recordingLogger{}
	manager := newFileManager(
		t,
		Opts{
			Logger: logger,
		},
	)
	failure := errors.New("nested failed")

	err := manager.Do(
		txcore.Flat(context.Background()),
		"root",
		func(ctx context.Context) error {
			if err := insert(
				ctx,
				manager,
				"root",
			); err != nil {
				return err
			}

			return manager.Do(
				ctx,
				"nested",
				func(ctx context.Context) error {
					return failure
				},
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
		ErrFlatNestedFailed,
	)
	require.Empty(
		t,
		logger.warns,
	)
	assertNames(
		t,
		manager,
	)
}

func TestFlatNestedSwallowedFailureRollsBackRoot(t *testing.T) {
	failure := errors.New("nested failed")

	tests := []struct {
		name       string
		nested     func(ctx context.Context, manager *TxManager) error
		wantErr    error
		wantErrors []string
	}{
		{
			name: "error",
			nested: func(ctx context.Context, manager *TxManager) error {
				if err := insert(
					ctx,
					manager,
					"nested",
				); err != nil {
					return err
				}

				return failure
			},
			wantErr: failure,
		},
		{
			name: "panic",
			nested: func(ctx context.Context, manager *TxManager) error {
				if err := insert(
					ctx,
					manager,
					"nested",
				); err != nil {
					return err
				}

				panic("boom")
			},
			wantErr:    ErrPanic,
			wantErrors: []string{"panic occurred during transaction"},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				logger := &recordingLogger{}
				manager := newFileManager(
					t,
					Opts{
						Logger: logger,
					},
				)
				log := &effectLog{}

				err := manager.Do(
					txcore.Flat(context.Background()),
					"root",
					func(ctx context.Context) error {
						require.True(
							t,
							txcore.AddEffect(
								ctx,
								log.effect("root"),
							),
						)

						nestedErr := manager.Do(
							ctx,
							"nested",
							func(ctx context.Context) error {
								return tt.nested(
									ctx,
									manager,
								)
							},
						)
						require.ErrorIs(
							t,
							nestedErr,
							tt.wantErr,
						)

						return insert(
							ctx,
							manager,
							"root",
						)
					},
				)

				require.ErrorIs(
					t,
					err,
					ErrFlatNestedFailed,
				)
				require.ErrorIs(
					t,
					err,
					tt.wantErr,
				)
				require.Equal(
					t,
					[]string{"flat nested transaction failed, root rolled back"},
					logger.warns,
				)
				require.Equal(
					t,
					tt.wantErrors,
					logger.errorMsg,
				)
				require.Empty(
					t,
					log.names(),
				)
				assertNames(
					t,
					manager,
				)
			},
		)
	}
}

func TestFlatNestedSharesRootDeadline(t *testing.T) {
	manager := newFileManager(
		t,
		Opts{
			Timeout: 2 * time.Second,
		},
	)

	err := manager.Do(
		context.Background(),
		"root",
		func(ctx context.Context) error {
			rootDeadline, ok := ctx.Deadline()
			require.True(
				t,
				ok,
			)

			time.Sleep(10 * time.Millisecond)

			require.NoError(
				t,
				manager.Do(
					txcore.Flat(ctx),
					"flat",
					func(ctx context.Context) error {
						flatDeadline, ok := ctx.Deadline()
						require.True(
							t,
							ok,
						)
						require.Equal(
							t,
							rootDeadline,
							flatDeadline,
						)

						return nil
					},
				),
			)

			return manager.Do(
				ctx,
				"savepoint",
				func(ctx context.Context) error {
					nestedDeadline, ok := ctx.Deadline()
					require.True(
						t,
						ok,
					)
					require.True(
						t,
						nestedDeadline.After(rootDeadline) || nestedDeadline.Equal(rootDeadline),
					)

					return nil
				},
			)
		},
	)

	require.NoError(
		t,
		err,
	)
}

func TestFlatNestedRetryIsRejectedWithoutMarking(t *testing.T) {
	manager := newTestManager(t)
	calls := 0

	err := manager.Do(
		txcore.Flat(context.Background()),
		"root",
		func(ctx context.Context) error {
			nestedErr := manager.Do(
				ctx,
				"nested",
				func(context.Context) error {
					calls++

					return nil
				},
				txcore.WithRetry(2),
			)
			require.ErrorIs(
				t,
				nestedErr,
				ErrRetryInNested,
			)

			return insert(
				ctx,
				manager,
				"root",
			)
		},
	)

	require.NoError(
		t,
		err,
	)
	require.Zero(
		t,
		calls,
	)
	assertNames(
		t,
		manager,
		"root",
	)
}

func TestWithSavepointUnderFlatRollsBackOnlyItsArea(t *testing.T) {
	logger := &recordingLogger{}
	manager := newFileManager(
		t,
		Opts{
			Logger: logger,
		},
	)
	log := recordSQL(
		t,
		manager,
	)
	effects := &effectLog{}
	failure := errors.New("isolated failed")

	err := manager.Do(
		txcore.Flat(context.Background()),
		"root",
		func(ctx context.Context) error {
			if err := insert(
				ctx,
				manager,
				"a",
			); err != nil {
				return err
			}

			require.True(
				t,
				txcore.AddEffect(
					ctx,
					effects.effect("a"),
				),
			)

			if err := manager.Do(
				ctx,
				"flat",
				func(ctx context.Context) error {
					return insert(
						ctx,
						manager,
						"b",
					)
				},
			); err != nil {
				return err
			}

			isolatedErr := manager.Do(
				ctx,
				"isolated",
				func(ctx context.Context) error {
					if err := insert(
						ctx,
						manager,
						"c",
					); err != nil {
						return err
					}

					require.True(
						t,
						txcore.AddEffect(
							ctx,
							effects.effect("c"),
						),
					)

					return failure
				},
				txcore.WithSavepoint(),
			)
			require.ErrorIs(
				t,
				isolatedErr,
				failure,
			)

			require.True(
				t,
				txcore.AddEffect(
					ctx,
					effects.effect("d"),
				),
			)

			return manager.Do(
				ctx,
				"flat-after",
				func(ctx context.Context) error {
					return insert(
						ctx,
						manager,
						"d",
					)
				},
			)
		},
	)

	require.NoError(
		t,
		err,
	)
	require.Equal(
		t,
		1,
		log.count(sqlSavepoint),
	)
	require.Equal(
		t,
		1,
		log.count(sqlRollbackTo),
	)
	require.Equal(
		t,
		[]string{"a", "d"},
		effects.names(),
	)
	require.Empty(
		t,
		logger.warns,
	)
	assertNames(
		t,
		manager,
		"a",
		"b",
		"d",
	)
}

func TestWithSavepointAbsorbsFlatFailureInside(t *testing.T) {
	logger := &recordingLogger{}
	manager := newFileManager(
		t,
		Opts{
			Logger: logger,
		},
	)
	effects := &effectLog{}
	failure := errors.New("flat failed")

	err := manager.Do(
		txcore.Flat(context.Background()),
		"root",
		func(ctx context.Context) error {
			isolatedErr := manager.Do(
				ctx,
				"isolated",
				func(ctx context.Context) error {
					if err := insert(
						ctx,
						manager,
						"s1",
					); err != nil {
						return err
					}

					require.True(
						t,
						txcore.AddEffect(
							ctx,
							effects.effect("s1"),
						),
					)

					flatErr := manager.Do(
						ctx,
						"flat",
						func(ctx context.Context) error {
							if err := insert(
								ctx,
								manager,
								"s2",
							); err != nil {
								return err
							}

							return failure
						},
					)
					require.ErrorIs(
						t,
						flatErr,
						failure,
					)

					return nil
				},
				txcore.WithSavepoint(),
			)
			require.ErrorIs(
				t,
				isolatedErr,
				ErrFlatNestedFailed,
			)
			require.ErrorIs(
				t,
				isolatedErr,
				failure,
			)

			require.True(
				t,
				txcore.AddEffect(
					ctx,
					effects.effect("root"),
				),
			)

			return insert(
				ctx,
				manager,
				"root",
			)
		},
	)

	require.NoError(
		t,
		err,
	)
	require.Equal(
		t,
		[]string{"flat nested transaction failed, savepoint rolled back"},
		logger.warns,
	)
	require.Equal(
		t,
		[]string{"root"},
		effects.names(),
	)
	assertNames(
		t,
		manager,
		"root",
	)
}

func TestWithSavepointOutsideFlat(t *testing.T) {
	tests := []struct {
		name    string
		nested  bool
		wantSQL []string
	}{
		{
			name:    "on root it is a no-op",
			wantSQL: []string{sqlInsert},
		},
		{
			name:    "nested without flat matches default",
			nested:  true,
			wantSQL: []string{sqlInsert, sqlSavepoint, sqlInsert},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				manager := newTestManager(t)
				log := recordSQL(
					t,
					manager,
				)

				err := manager.Do(
					context.Background(),
					"root",
					func(ctx context.Context) error {
						if err := insert(
							ctx,
							manager,
							"root",
						); err != nil {
							return err
						}

						if !tt.nested {
							return nil
						}

						return manager.Do(
							ctx,
							"nested",
							func(ctx context.Context) error {
								return insert(
									ctx,
									manager,
									"nested",
								)
							},
							txcore.WithSavepoint(),
						)
					},
					txcore.WithSavepoint(),
				)

				require.NoError(
					t,
					err,
				)
				require.Equal(
					t,
					tt.wantSQL,
					log.kinds(),
				)
			},
		)
	}
}

func TestDetachInsideFlatOpensOwnRoot(t *testing.T) {
	manager := newFileManager(
		t,
		Opts{},
	)
	log := recordSQL(
		t,
		manager,
	)
	failure := errors.New("outer failed")

	err := manager.Do(
		txcore.Flat(context.Background()),
		"outer",
		func(ctx context.Context) error {
			detached := txcore.Detach(ctx)
			require.False(
				t,
				txcore.IsFlat(detached),
			)
			require.False(
				t,
				manager.InTransaction(detached),
			)

			if err := manager.Do(
				detached,
				"detached",
				func(ctx context.Context) error {
					if err := insert(
						ctx,
						manager,
						"detached",
					); err != nil {
						return err
					}

					return manager.Do(
						ctx,
						"detached-nested",
						func(ctx context.Context) error {
							return insert(
								ctx,
								manager,
								"detached-nested",
							)
						},
					)
				},
			); err != nil {
				return err
			}

			if err := insert(
				ctx,
				manager,
				"outer",
			); err != nil {
				return err
			}

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
		1,
		log.count(sqlSavepoint),
	)
	assertNames(
		t,
		manager,
		"detached",
		"detached-nested",
	)
}
