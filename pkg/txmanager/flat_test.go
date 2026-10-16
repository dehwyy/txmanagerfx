package txmanager_test

import (
	"context"
	"errors"
	"testing"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager"
	"github.com/dehwyy/txmanagerfx/pkg/txmanager/gormtx"
	"github.com/stretchr/testify/require"
)

func TestFlatNestedFailedIsShared(t *testing.T) {
	require.Same(
		t,
		gormtx.ErrFlatNestedFailed,
		txmanager.ErrFlatNestedFailed,
	)
}

func TestFlatMarkerOnPlainContext(t *testing.T) {
	ctx := txmanager.Flat(context.Background())

	require.False(
		t,
		txmanager.InTransaction(ctx),
	)
	require.False(
		t,
		txmanager.AfterCommit(
			ctx,
			func(context.Context) error { return nil },
		),
	)
}

func TestFlatSwallowedNestedFailure(t *testing.T) {
	manager, err := gormtx.New(gormtx.Opts{
		DB: newDB(t),
	})
	require.NoError(
		t,
		err,
	)

	failure := errors.New("nested failed")

	err = manager.Do(
		txmanager.Flat(context.Background()),
		"root",
		func(ctx context.Context) error {
			require.ErrorIs(
				t,
				manager.Do(
					ctx,
					"nested",
					func(context.Context) error { return failure },
				),
				failure,
			)

			return manager.Do(
				ctx,
				"isolated",
				func(context.Context) error { return nil },
				txmanager.WithSavepoint(),
			)
		},
	)

	require.ErrorIs(
		t,
		err,
		txmanager.ErrFlatNestedFailed,
	)
	require.ErrorIs(
		t,
		err,
		failure,
	)
}
