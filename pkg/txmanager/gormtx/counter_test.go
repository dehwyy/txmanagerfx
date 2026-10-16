package gormtx

import (
	"context"
	"fmt"
	"testing"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
	"github.com/stretchr/testify/require"
)

func TestSubtransactionCounterWarnsOncePerAttempt(t *testing.T) {
	const warning = "subtransaction count exceeded 64"

	tests := []struct {
		name      string
		nested    int
		flat      bool
		depth     bool
		attempts  int
		wantWarns int
	}{
		{
			name:   "64 sequential savepoints",
			nested: 64,
		},
		{
			name:      "65 sequential savepoints",
			nested:    65,
			wantWarns: 1,
		},
		{
			name:      "130 sequential savepoints",
			nested:    130,
			wantWarns: 1,
		},
		{
			name:      "65 savepoints by depth",
			nested:    65,
			depth:     true,
			wantWarns: 1,
		},
		{
			name:   "flat nested issue no savepoints",
			nested: 130,
			flat:   true,
		},
		{
			name:      "each retry attempt warns once",
			nested:    65,
			attempts:  2,
			wantWarns: 2,
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
				ctx := context.Background()

				if tt.flat {
					ctx = txcore.Flat(ctx)
				}

				calls := 0

				err := manager.Do(
					ctx,
					"root",
					func(ctx context.Context) error {
						calls++

						if tt.depth {
							if err := nestDeep(
								ctx,
								manager,
								tt.nested,
							); err != nil {
								return err
							}
						} else {
							for i := range tt.nested {
								if err := manager.Do(
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
						}

						if calls < tt.attempts {
							return &sqlStateError{
								code: "40001",
							}
						}

						return nil
					},
					txcore.WithRetry(max(
						tt.attempts,
						1,
					)),
					txcore.WithBackoff(noBackoff),
				)

				require.NoError(
					t,
					err,
				)

				found := 0

				for _, msg := range logger.warns {
					if msg == warning {
						found++
					}
				}

				require.Equal(
					t,
					tt.wantWarns,
					found,
				)
			},
		)
	}
}

func nestDeep(ctx context.Context, manager *TxManager, levels int) error {
	if levels == 0 {
		return nil
	}

	return manager.Do(
		ctx,
		fmt.Sprintf(
			"d%d",
			levels,
		),
		func(ctx context.Context) error {
			return nestDeep(
				ctx,
				manager,
				levels-1,
			)
		},
	)
}
