package gormtx

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
	"github.com/stretchr/testify/require"
)

func TestLoggerReceivesPanicAndRetry(t *testing.T) {
	tests := []struct {
		name      string
		run       func(manager *TxManager) error
		wantWarn  []string
		wantError []string
		wantFail  bool
	}{
		{
			name: "panic is logged as error",
			run: func(manager *TxManager) error {
				return manager.Do(context.Background(), "p", func(ctx context.Context) error {
					panic("boom")
				})
			},
			wantError: []string{"panic occurred during transaction"},
			wantFail:  true,
		},
		{
			name: "retry is logged as warn",
			run: func(manager *TxManager) error {
				calls := 0

				return manager.Do(
					context.Background(),
					"r",
					func(ctx context.Context) error {
						calls++
						if calls == 1 {
							return &sqlStateError{code: "40P01"}
						}

						return nil
					},
					txcore.WithRetry(2),
					txcore.WithBackoff(noBackoff),
				)
			},
			wantWarn: []string{"transaction retry"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := &recordingLogger{}
			manager := newFileManager(t, Opts{Logger: logger})

			err := tt.run(manager)
			if tt.wantFail {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.wantWarn, logger.warns)
			require.Equal(t, tt.wantError, logger.errorMsg)
		})
	}
}

func TestDefaultLoggerIsSilentAndSafe(t *testing.T) {
	manager := newTestManager(t)

	err := manager.Do(context.Background(), "p", func(ctx context.Context) error {
		panic("boom")
	})

	require.ErrorIs(t, err, ErrPanic)
}

func TestTimeoutOption(t *testing.T) {
	tests := []struct {
		name         string
		timeout      time.Duration
		wantDeadline bool
		wantAtMost   time.Duration
	}{
		{
			name:         "custom timeout applied",
			timeout:      2 * time.Second,
			wantDeadline: true,
			wantAtMost:   2 * time.Second,
		},
		{
			name:         "zero falls back to default",
			timeout:      0,
			wantDeadline: true,
			wantAtMost:   defaultTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := newFileManager(t, Opts{Timeout: tt.timeout})

			err := manager.Do(context.Background(), "t", func(ctx context.Context) error {
				deadline, ok := ctx.Deadline()
				require.Equal(t, tt.wantDeadline, ok)
				require.LessOrEqual(t, time.Until(deadline), tt.wantAtMost)
				require.Greater(t, time.Until(deadline), tt.wantAtMost-time.Second)

				return nil
			})

			require.NoError(t, err)
		})
	}
}

func TestCancelledContextIsLogged(t *testing.T) {
	logger := &recordingLogger{}
	manager := newFileManager(t, Opts{Logger: logger})
	ctx, cancel := context.WithCancel(context.Background())

	err := manager.Do(ctx, "c", func(ctx context.Context) error {
		cancel()

		return errors.New("failed")
	})

	require.Error(t, err)
	require.Contains(t, logger.warns, "transaction context timeout/cancelled")
}
