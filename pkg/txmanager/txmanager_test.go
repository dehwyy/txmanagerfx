package txmanager_test

import (
	"context"
	"testing"
	"time"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager"
	"github.com/dehwyy/txmanagerfx/pkg/txmanager/gormtx"
	"github.com/dehwyy/txmanagerfx/pkg/txmanager/txmanagertest"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

func newDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(
		sqlite.Open("file::memory:"),
		&gorm.Config{},
	)
	require.NoError(t, err)

	return db
}

func TestErrorsAreShared(t *testing.T) {
	require.Same(t, gormtx.ErrPanic, txmanager.ErrPanic)
	require.Same(t, gormtx.ErrRetryInNested, txmanager.ErrRetryInNested)
}

func TestFreeFunctionsOnPlainContext(t *testing.T) {
	ctx := context.Background()

	require.False(t, txmanager.InTransaction(ctx))
	require.False(t, txmanager.InTransaction(txmanager.Detach(ctx)))
}

func TestNewGormGraph(t *testing.T) {
	tests := []struct {
		name    string
		options []fx.Option
	}{
		{
			name: "without logger and timeout",
			options: []fx.Option{
				fx.Provide(func() *gorm.DB { return newDB(t) }),
			},
		},
		{
			name: "with logger and timeout",
			options: []fx.Option{
				fx.Provide(func() *gorm.DB { return newDB(t) }),
				fx.Provide(func() gormtx.Logger { return noopLogger{} }),
				fx.Provide(func() time.Duration { return time.Second }),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var manager txmanager.TxManager

			app := fx.New(
				append(
					tt.options,
					txmanager.NewGorm(),
					fx.Populate(&manager),
					fx.NopLogger,
				)...,
			)
			require.NoError(t, app.Err())
			require.NotNil(t, manager)

			err := manager.Do(context.Background(), "t", func(ctx context.Context) error {
				require.True(t, manager.InTransaction(ctx))

				return nil
			})
			require.NoError(t, err)
		})
	}
}

type noopLogger struct{}

func (noopLogger) Warn(string, ...any) {}

func (noopLogger) Error(string, ...any) {}

func TestOptionsAcceptedByFake(t *testing.T) {
	var manager txmanager.TxManager = txmanagertest.New()

	err := manager.Do(
		context.Background(),
		"t",
		func(ctx context.Context) error { return nil },
		txmanager.WithRetry(3),
		txmanager.WithBackoff(func(int) time.Duration { return 0 }),
	)

	require.NoError(t, err)
}
