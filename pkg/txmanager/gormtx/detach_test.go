package gormtx

import (
	"context"
	"errors"
	"testing"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
	"github.com/stretchr/testify/require"
)

func TestDetachWriteSurvivesOuterRollback(t *testing.T) {
	tests := []struct {
		name      string
		outerErr  error
		wantNames []string
	}{
		{
			name:      "outer commits",
			outerErr:  nil,
			wantNames: []string{"detached", "tx"},
		},
		{
			name:      "outer rolls back",
			outerErr:  errors.New("outer failed"),
			wantNames: []string{"detached"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := newFileManager(t, Opts{})

			err := manager.Do(context.Background(), "outer", func(ctx context.Context) error {
				if err := manager.Do(txcore.Detach(ctx), "detached", func(ctx context.Context) error {
					return insert(ctx, manager, "detached")
				}); err != nil {
					return err
				}

				if tt.outerErr == nil {
					if err := insert(ctx, manager, "tx"); err != nil {
						return err
					}
				}

				return tt.outerErr
			})

			if tt.outerErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.outerErr)
			}

			assertNames(t, manager, tt.wantNames...)
		})
	}
}

func TestDetachGetConnectionIsRaw(t *testing.T) {
	manager := newFileManager(t, Opts{})

	err := manager.Do(context.Background(), "outer", func(ctx context.Context) error {
		require.NoError(t, insert(txcore.Detach(ctx), manager, "detached"))

		return errors.New("rollback outer")
	})
	require.Error(t, err)

	assertNames(t, manager, "detached")
}

func TestDetachDoOpensNewRootTransaction(t *testing.T) {
	manager := newFileManager(t, Opts{})

	err := manager.Do(context.Background(), "outer", func(ctx context.Context) error {
		detached := txcore.Detach(ctx)

		return manager.Do(
			detached,
			"inner-root",
			func(inner context.Context) error {
				require.True(t, manager.InTransaction(inner))
				require.NoError(t, insert(inner, manager, "inner"))

				return errors.New("inner rollback")
			},
			txcore.WithRetry(2),
		)
	})
	require.Error(t, err)

	assertNames(t, manager)
}

func TestDetachDoesNotMutateParentContext(t *testing.T) {
	manager := newTestManager(t)

	err := manager.Do(context.Background(), "outer", func(ctx context.Context) error {
		detached := txcore.Detach(ctx)
		require.False(t, manager.InTransaction(detached))

		require.True(t, manager.InTransaction(ctx))

		return insert(ctx, manager, "still-in-tx")
	})
	require.NoError(t, err)

	assertNames(t, manager, "still-in-tx")
}
