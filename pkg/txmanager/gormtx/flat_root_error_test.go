package gormtx

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
	"github.com/stretchr/testify/require"
)

func TestFlatSwallowedFailureThenRootError(t *testing.T) {
	failure := errors.New("nested failed")
	aborted := &sqlStateError{
		code: "25P02",
	}

	tests := []struct {
		name      string
		rootErr   func(nestedErr error) error
		wantFlat  bool
		wantWarns []string
		wantState string
	}{
		{
			name: "root returns another error",
			rootErr: func(error) error {
				return aborted
			},
			wantFlat:  true,
			wantWarns: []string{"flat nested transaction failed, root rolled back"},
			wantState: "25P02",
		},
		{
			name: "root wraps the nested error",
			rootErr: func(nestedErr error) error {
				return fmt.Errorf(
					"apply: %w",
					nestedErr,
				)
			},
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

				err := manager.Do(
					txcore.Flat(context.Background()),
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
						require.ErrorIs(
							t,
							nestedErr,
							failure,
						)

						return tt.rootErr(nestedErr)
					},
				)

				require.ErrorIs(
					t,
					err,
					failure,
				)
				require.Equal(
					t,
					tt.wantFlat,
					errors.Is(
						err,
						ErrFlatNestedFailed,
					),
				)
				require.Equal(
					t,
					tt.wantWarns,
					logger.warns,
				)
				require.False(
					t,
					txcore.Retryable(err),
				)

				if tt.wantState != "" {
					require.ErrorIs(
						t,
						err,
						aborted,
					)

					var stater interface {
						SQLState() string
					}
					require.ErrorAs(
						t,
						err,
						&stater,
					)
					require.Equal(
						t,
						tt.wantState,
						stater.SQLState(),
					)
				}

				assertNames(
					t,
					manager,
				)
			},
		)
	}
}

func TestFlatSwallowedFailureThenRetryableRootError(t *testing.T) {
	logger := &recordingLogger{}
	manager := newFileManager(
		t,
		Opts{
			Logger: logger,
		},
	)
	failure := errors.New("nested failed")
	calls := 0

	err := manager.Do(
		txcore.Flat(context.Background()),
		"root",
		func(ctx context.Context) error {
			calls++

			if calls == 1 {
				require.ErrorIs(
					t,
					manager.Do(
						ctx,
						"nested",
						func(context.Context) error {
							return failure
						},
					),
					failure,
				)

				return &sqlStateError{
					code: "40001",
				}
			}

			return insert(
				ctx,
				manager,
				"retried",
			)
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
		[]string{"flat nested transaction failed, root rolled back", "transaction retry"},
		logger.warns,
	)
	assertNames(
		t,
		manager,
		"retried",
	)
}

func TestWithSavepointAreaReturnsAnotherErrorAfterFlatFailure(t *testing.T) {
	logger := &recordingLogger{}
	manager := newFileManager(
		t,
		Opts{
			Logger: logger,
		},
	)
	failure := errors.New("flat failed")
	aborted := &sqlStateError{
		code: "25P02",
	}

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
						"isolated",
					); err != nil {
						return err
					}

					require.ErrorIs(
						t,
						manager.Do(
							ctx,
							"flat",
							func(context.Context) error {
								return failure
							},
						),
						failure,
					)

					return aborted
				},
				txcore.WithSavepoint(),
			)
			require.ErrorIs(
				t,
				isolatedErr,
				aborted,
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
		[]string{"savepoint area rolled back after flat nested failure"},
		logger.warns,
	)
	assertNames(
		t,
		manager,
		"root",
	)
}
