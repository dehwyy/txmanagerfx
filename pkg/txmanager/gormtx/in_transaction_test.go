package gormtx

import (
	"context"
	"testing"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
	"github.com/stretchr/testify/require"
)

func TestInTransaction(t *testing.T) {
	manager := newTestManager(t)

	tests := []struct {
		name string
		run  func(t *testing.T) bool
		want bool
	}{
		{
			name: "outside any transaction",
			run: func(t *testing.T) bool {
				return manager.InTransaction(context.Background())
			},
			want: false,
		},
		{
			name: "inside Do",
			run: func(t *testing.T) bool {
				var got bool

				err := manager.Do(context.Background(), "root", func(ctx context.Context) error {
					got = manager.InTransaction(ctx)

					return nil
				})
				require.NoError(t, err)

				return got
			},
			want: true,
		},
		{
			name: "inside nested Do",
			run: func(t *testing.T) bool {
				var got bool

				err := manager.Do(context.Background(), "root", func(ctx context.Context) error {
					return manager.Do(ctx, "nested", func(ctx context.Context) error {
						got = manager.InTransaction(ctx)

						return nil
					})
				})
				require.NoError(t, err)

				return got
			},
			want: true,
		},
		{
			name: "inside detached context",
			run: func(t *testing.T) bool {
				var got bool

				err := manager.Do(context.Background(), "root", func(ctx context.Context) error {
					got = manager.InTransaction(txcore.Detach(ctx))

					return nil
				})
				require.NoError(t, err)

				return got
			},
			want: false,
		},
		{
			name: "free function matches the method",
			run: func(t *testing.T) bool {
				var got bool

				err := manager.Do(context.Background(), "root", func(ctx context.Context) error {
					got = txcore.Active(ctx)

					return nil
				})
				require.NoError(t, err)

				return got
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.run(t))
		})
	}
}
