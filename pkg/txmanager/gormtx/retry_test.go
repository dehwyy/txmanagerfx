package gormtx

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
	"github.com/stretchr/testify/require"
)

func TestRetryRootTransaction(t *testing.T) {
	plain := errors.New("plain failure")

	tests := []struct {
		name         string
		attempts     int
		failures     []error
		wantCalls    int
		wantErr      error
		wantAnyError bool
		wantNames    []string
	}{
		{
			name:      "deadlock then success",
			attempts:  3,
			failures:  []error{&sqlStateError{code: "40P01"}},
			wantCalls: 2,
			wantNames: []string{"attempt-2"},
		},
		{
			name:      "serialization failure then success",
			attempts:  3,
			failures:  []error{&sqlStateError{code: "40001"}, &sqlStateError{code: "40001"}},
			wantCalls: 3,
			wantNames: []string{"attempt-3"},
		},
		{
			name:         "attempts exhausted returns last error",
			attempts:     3,
			failures:     []error{&sqlStateError{code: "40P01"}, &sqlStateError{code: "40P01"}, &sqlStateError{code: "40P01"}},
			wantCalls:    3,
			wantAnyError: true,
		},
		{
			name:         "non retryable sqlstate is not retried",
			attempts:     3,
			failures:     []error{&sqlStateError{code: "23505"}},
			wantCalls:    1,
			wantErr:      nil,
			wantNames:    nil,
			wantAnyError: true,
		},
		{
			name:      "plain error is not retried",
			attempts:  3,
			failures:  []error{plain},
			wantCalls: 1,
			wantErr:   plain,
		},
		{
			name:      "wrapped retryable error is retried",
			attempts:  2,
			failures:  []error{errors.Join(errors.New("ctx"), &sqlStateError{code: "40001"})},
			wantCalls: 2,
			wantNames: []string{"attempt-2"},
		},
		{
			name:         "zero attempts means a single try",
			attempts:     0,
			failures:     []error{&sqlStateError{code: "40P01"}},
			wantCalls:    1,
			wantAnyError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := newTestManager(t)
			calls := 0

			err := manager.Do(
				context.Background(),
				"retry",
				func(ctx context.Context) error {
					calls++

					if err := insert(ctx, manager, "attempt-"+string(rune('0'+calls))); err != nil {
						return err
					}

					if calls <= len(tt.failures) {
						return tt.failures[calls-1]
					}

					return nil
				},
				txcore.WithRetry(tt.attempts),
				txcore.WithBackoff(noBackoff),
			)

			require.Equal(t, tt.wantCalls, calls)

			switch {
			case tt.wantErr != nil:
				require.ErrorIs(t, err, tt.wantErr)
			case tt.wantAnyError:
				require.Error(t, err)
			default:
				require.NoError(t, err)
			}

			assertNames(t, manager, tt.wantNames...)
		})
	}
}

func TestWithoutRetryOptionRunsOnce(t *testing.T) {
	manager := newTestManager(t)
	calls := 0

	err := manager.Do(context.Background(), "once", func(ctx context.Context) error {
		calls++

		return &sqlStateError{code: "40P01"}
	})

	require.Error(t, err)
	require.Equal(t, 1, calls)
}

func TestRetryDoesNotRetryPanic(t *testing.T) {
	manager := newTestManager(t)
	calls := 0

	err := manager.Do(
		context.Background(),
		"panicking",
		func(ctx context.Context) error {
			calls++

			panic("boom")
		},
		txcore.WithRetry(3),
		txcore.WithBackoff(noBackoff),
	)

	require.ErrorIs(t, err, ErrPanic)
	require.Equal(t, 1, calls)
}

func TestRetryInNestedIsRejected(t *testing.T) {
	manager := newTestManager(t)
	innerCalls := 0

	err := manager.Do(context.Background(), "outer", func(ctx context.Context) error {
		innerErr := manager.Do(
			ctx,
			"inner",
			func(ctx context.Context) error {
				innerCalls++

				return nil
			},
			txcore.WithRetry(3),
		)
		require.ErrorIs(t, innerErr, ErrRetryInNested)

		return insert(ctx, manager, "outer")
	})

	require.NoError(t, err)
	require.Zero(t, innerCalls)
	assertNames(t, manager, "outer")
}

func TestRetryStopsWhenContextCancelledDuringBackoff(t *testing.T) {
	manager := newTestManager(t)
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0

	err := manager.Do(
		ctx,
		"cancel",
		func(ctx context.Context) error {
			calls++
			cancel()

			return &sqlStateError{code: "40P01"}
		},
		txcore.WithRetry(5),
		txcore.WithBackoff(func(int) time.Duration { return time.Hour }),
	)

	var stateErr *sqlStateError

	require.ErrorAs(t, err, &stateErr)
	require.Equal(t, 1, calls)
}

func TestRetryUsesBackoffPerAttempt(t *testing.T) {
	manager := newTestManager(t)

	var attempts []int

	err := manager.Do(
		context.Background(),
		"backoff",
		func(ctx context.Context) error {
			return &sqlStateError{code: "40001"}
		},
		txcore.WithRetry(3),
		txcore.WithBackoff(func(attempt int) time.Duration {
			attempts = append(attempts, attempt)

			return 0
		}),
	)

	require.Error(t, err)
	require.Equal(t, []int{1, 2}, attempts)
}
