package gormtx

import (
	"context"
	"errors"
	"testing"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
	"github.com/stretchr/testify/require"
)

func TestFlatFailureSwallowedTwoLevelsDeepRollsBackRoot(t *testing.T) {
	logger := &recordingLogger{}
	manager := newFileManager(
		t,
		Opts{
			Logger: logger,
		},
	)
	failure := errors.New("deep failed")

	err := manager.Do(
		txcore.Flat(context.Background()),
		"root",
		func(ctx context.Context) error {
			if err := insert(
				ctx,
				manager,
				"root",
			); err != nil {
				return err
			}

			return manager.Do(
				ctx,
				"middle",
				func(ctx context.Context) error {
					if err := insert(
						ctx,
						manager,
						"middle",
					); err != nil {
						return err
					}

					require.ErrorIs(
						t,
						manager.Do(
							ctx,
							"deep",
							func(context.Context) error { return failure },
						),
						failure,
					)

					return nil
				},
			)
		},
	)

	require.ErrorIs(
		t,
		err,
		ErrFlatNestedFailed,
	)
	require.ErrorIs(
		t,
		err,
		failure,
	)
	require.Equal(
		t,
		[]string{"flat nested transaction failed, root rolled back"},
		logger.warns,
	)
	assertNames(
		t,
		manager,
	)
}
