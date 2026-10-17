package txmanagertest

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager"
	"github.com/stretchr/testify/require"
)

func TestFakeFlatSwallowedFailureThenRootError(t *testing.T) {
	failure := errors.New("nested failed")
	aborted := errors.New("current transaction is aborted")

	tests := []struct {
		name      string
		rootErr   func(nestedErr error) error
		wantFlat  bool
		wantWarns []string
	}{
		{
			name: "root returns another error",
			rootErr: func(error) error {
				return aborted
			},
			wantFlat:  true,
			wantWarns: []string{"flat nested transaction failed, root rolled back"},
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
				fake := &Fake{
					Logger: logger,
				}
				log := &effectLog{}

				err := fake.Do(
					txmanager.Flat(context.Background()),
					"root",
					func(ctx context.Context) error {
						require.True(
							t,
							txmanager.AfterCommit(
								ctx,
								log.effect("root"),
							),
						)

						nestedErr := fake.Do(
							ctx,
							"nested",
							func(context.Context) error { return failure },
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
						txmanager.ErrFlatNestedFailed,
					),
				)
				require.Equal(
					t,
					tt.wantFlat,
					errors.Is(
						err,
						aborted,
					),
				)
				require.Equal(
					t,
					tt.wantWarns,
					logger.warns,
				)
				require.Empty(
					t,
					log.names(),
				)
			},
		)
	}
}

func TestFakeWithSavepointAreaReturnsAnotherErrorAfterFlatFailure(t *testing.T) {
	logger := &recordingLogger{}
	fake := &Fake{
		Logger: logger,
	}
	log := &effectLog{}
	failure := errors.New("flat failed")
	aborted := errors.New("current transaction is aborted")

	err := fake.Do(
		txmanager.Flat(context.Background()),
		"root",
		func(ctx context.Context) error {
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
					require.ErrorIs(
						t,
						fake.Do(
							ctx,
							"flat",
							func(context.Context) error { return failure },
						),
						failure,
					)

					return aborted
				},
				txmanager.WithSavepoint(),
			)
			require.ErrorIs(
				t,
				isolatedErr,
				aborted,
			)
			require.ErrorIs(
				t,
				isolatedErr,
				txmanager.ErrFlatNestedFailed,
			)
			require.ErrorIs(
				t,
				isolatedErr,
				failure,
			)

			require.True(
				t,
				txmanager.AfterCommit(
					ctx,
					log.effect("root"),
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
		[]string{"savepoint area rolled back after flat nested failure"},
		logger.warns,
	)
	require.Equal(
		t,
		[]string{"root"},
		log.names(),
	)
}
