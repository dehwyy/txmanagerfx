package txmanagertest

import (
	"context"
	"errors"
	"testing"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager"
	"github.com/stretchr/testify/require"
)

func TestFakeDo(t *testing.T) {
	boom := errors.New("boom")

	tests := []struct {
		name      string
		fake      *Fake
		fn        func(ctx context.Context) error
		wantErr   error
		wantRuns  int
		wantCalls []Call
	}{
		{
			name:      "runs fn and counts the call",
			fake:      New(),
			fn:        func(context.Context) error { return nil },
			wantRuns:  1,
			wantCalls: []Call{{Name: "tx"}},
		},
		{
			name:      "propagates fn error",
			fake:      New(),
			fn:        func(context.Context) error { return boom },
			wantErr:   boom,
			wantRuns:  1,
			wantCalls: []Call{{Name: "tx"}},
		},
		{
			name:      "injected error skips fn",
			fake:      &Fake{DoErr: boom},
			fn:        func(context.Context) error { return nil },
			wantErr:   boom,
			wantRuns:  0,
			wantCalls: []Call{{Name: "tx"}},
		},
		{
			name:      "commit error after fn",
			fake:      &Fake{CommitErr: boom},
			fn:        func(context.Context) error { return nil },
			wantErr:   boom,
			wantRuns:  1,
			wantCalls: []Call{{Name: "tx"}},
		},
		{
			name:      "panic becomes ErrPanic",
			fake:      New(),
			fn:        func(context.Context) error { panic("x") },
			wantErr:   txmanager.ErrPanic,
			wantRuns:  1,
			wantCalls: []Call{{Name: "tx"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runs := 0

			err := tt.fake.Do(
				context.Background(),
				"tx",
				func(ctx context.Context) error {
					runs++

					return tt.fn(ctx)
				},
			)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.wantRuns, runs)
			require.Equal(t, tt.wantCalls, tt.fake.Calls())
			require.Equal(t, len(tt.wantCalls), tt.fake.CallCount())
		})
	}
}

func TestFakeInTransactionAndNesting(t *testing.T) {
	fake := New()
	ctx := context.Background()

	require.False(t, fake.InTransaction(ctx))

	err := fake.Do(ctx, "outer", func(ctx context.Context) error {
		require.True(t, fake.InTransaction(ctx))
		require.True(t, txmanager.InTransaction(ctx))
		require.False(t, fake.InTransaction(txmanager.Detach(ctx)))

		return fake.Do(ctx, "inner", func(ctx context.Context) error {
			return nil
		})
	})

	require.NoError(t, err)
	require.Equal(
		t,
		[]Call{
			{Name: "outer"},
			{Name: "inner", Nested: true},
		},
		fake.Calls(),
	)
}

func TestFakeRetryInNested(t *testing.T) {
	fake := New()

	err := fake.Do(context.Background(), "outer", func(ctx context.Context) error {
		return fake.Do(
			ctx,
			"inner",
			func(context.Context) error { return nil },
			txmanager.WithRetry(2),
		)
	})

	require.ErrorIs(t, err, txmanager.ErrRetryInNested)
}

func TestFakeConnections(t *testing.T) {
	require.Nil(t, New().GetConnection(context.Background()))
	require.Nil(t, New().GetRawConnection(context.Background()))
}

func TestFakeReset(t *testing.T) {
	fake := New()

	require.NoError(t, fake.Do(context.Background(), "a", func(context.Context) error { return nil }))
	fake.Reset()

	require.Zero(t, fake.CallCount())
	require.Empty(t, fake.Calls())
}
