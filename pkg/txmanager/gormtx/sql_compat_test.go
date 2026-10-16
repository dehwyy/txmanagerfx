package gormtx

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSQLWithoutNewOptionsMatchesV010(t *testing.T) {
	failure := errors.New("inner failed")

	tests := []struct {
		name      string
		run       func(ctx context.Context, manager *TxManager) error
		wantSQL   []string
		wantNames []string
	}{
		{
			name: "root",
			run: func(ctx context.Context, manager *TxManager) error {
				return manager.Do(
					ctx,
					"root",
					func(ctx context.Context) error {
						return insert(
							ctx,
							manager,
							"root",
						)
					},
				)
			},
			wantSQL:   []string{sqlInsert},
			wantNames: []string{"root"},
		},
		{
			name: "nested commits",
			run: func(ctx context.Context, manager *TxManager) error {
				return manager.Do(
					ctx,
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
								return insert(
									ctx,
									manager,
									"nested",
								)
							},
						)
					},
				)
			},
			wantSQL:   []string{sqlInsert, sqlSavepoint, sqlInsert},
			wantNames: []string{"root", "nested"},
		},
		{
			name: "nested error swallowed",
			run: func(ctx context.Context, manager *TxManager) error {
				return manager.Do(
					ctx,
					"root",
					func(ctx context.Context) error {
						nestedErr := manager.Do(
							ctx,
							"nested",
							func(ctx context.Context) error {
								if err := insert(
									ctx,
									manager,
									"nested",
								); err != nil {
									return err
								}

								return failure
							},
						)
						if !errors.Is(
							nestedErr,
							failure,
						) {
							return errors.New("unexpected nested result")
						}

						return insert(
							ctx,
							manager,
							"root",
						)
					},
				)
			},
			wantSQL:   []string{sqlSavepoint, sqlInsert, sqlRollbackTo, sqlInsert},
			wantNames: []string{"root"},
		},
		{
			name: "nested panic swallowed",
			run: func(ctx context.Context, manager *TxManager) error {
				return manager.Do(
					ctx,
					"root",
					func(ctx context.Context) error {
						nestedErr := manager.Do(
							ctx,
							"nested",
							func(ctx context.Context) error {
								if err := insert(
									ctx,
									manager,
									"nested",
								); err != nil {
									return err
								}

								panic("boom")
							},
						)
						if !errors.Is(
							nestedErr,
							ErrPanic,
						) {
							return errors.New("unexpected nested result")
						}

						return insert(
							ctx,
							manager,
							"root",
						)
					},
				)
			},
			wantSQL:   []string{sqlSavepoint, sqlInsert, sqlRollbackTo, sqlInsert},
			wantNames: []string{"root"},
		},
		{
			name: "three levels",
			run: func(ctx context.Context, manager *TxManager) error {
				return manager.Do(
					ctx,
					"l1",
					func(ctx context.Context) error {
						return manager.Do(
							ctx,
							"l2",
							func(ctx context.Context) error {
								if err := insert(
									ctx,
									manager,
									"l2",
								); err != nil {
									return err
								}

								l3Err := manager.Do(
									ctx,
									"l3",
									func(ctx context.Context) error {
										if err := insert(
											ctx,
											manager,
											"l3",
										); err != nil {
											return err
										}

										return failure
									},
								)
								if !errors.Is(
									l3Err,
									failure,
								) {
									return errors.New("unexpected l3 result")
								}

								return insert(
									ctx,
									manager,
									"l2-after",
								)
							},
						)
					},
				)
			},
			wantSQL: []string{
				sqlSavepoint,
				sqlInsert,
				sqlSavepoint,
				sqlInsert,
				sqlRollbackTo,
				sqlInsert,
			},
			wantNames: []string{"l2", "l2-after"},
		},
		{
			name: "root error",
			run: func(ctx context.Context, manager *TxManager) error {
				err := manager.Do(
					ctx,
					"root",
					func(ctx context.Context) error {
						if err := insert(
							ctx,
							manager,
							"root",
						); err != nil {
							return err
						}

						return failure
					},
				)
				if !errors.Is(
					err,
					failure,
				) {
					return errors.New("unexpected root result")
				}

				return nil
			},
			wantSQL: []string{sqlInsert},
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

				require.NoError(
					t,
					tt.run(
						context.Background(),
						manager,
					),
				)
				require.Equal(
					t,
					tt.wantSQL,
					log.kinds(),
				)

				assertNames(
					t,
					manager,
					tt.wantNames...,
				)
			},
		)
	}
}
