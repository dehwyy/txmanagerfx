package txmanager_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager"
	"github.com/dehwyy/txmanagerfx/pkg/txmanager/gormtx"
	"github.com/stretchr/testify/require"
)

func TestAfterCommit(t *testing.T) {
	failure := errors.New("failed")

	tests := []struct {
		name    string
		fnErr   error
		wantRan bool
	}{
		{
			name:    "runs after commit",
			wantRan: true,
		},
		{
			name:  "dropped on rollback",
			fnErr: failure,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				manager, err := gormtx.New(gormtx.Opts{
					DB: newDB(t),
				})
				require.NoError(
					t,
					err,
				)

				ran := false
				effect := func(ctx context.Context) error {
					require.False(
						t,
						txmanager.InTransaction(ctx),
					)

					ran = true

					return nil
				}

				require.False(
					t,
					txmanager.AfterCommit(
						context.Background(),
						effect,
					),
				)

				err = manager.Do(
					context.Background(),
					"t",
					func(ctx context.Context) error {
						require.True(
							t,
							txmanager.AfterCommit(
								ctx,
								effect,
							),
						)
						require.False(
							t,
							ran,
						)

						return tt.fnErr
					},
				)

				require.ErrorIs(
					t,
					err,
					tt.fnErr,
				)
				require.Equal(
					t,
					tt.wantRan,
					ran,
				)
			},
		)
	}
}
