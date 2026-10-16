package txmanagertest

import (
	"context"
	"errors"
	"testing"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager"
	"github.com/stretchr/testify/require"
)

func TestFakeDoErrOnFlatNestedMarksArea(t *testing.T) {
	fake := New()
	failure := errors.New("injected")
	runs := 0

	err := fake.Do(
		txmanager.Flat(context.Background()),
		"root",
		func(ctx context.Context) error {
			fake.DoErr = failure

			nestedErr := fake.Do(
				ctx,
				"nested",
				func(context.Context) error {
					runs++

					return nil
				},
			)

			fake.DoErr = nil

			require.ErrorIs(
				t,
				nestedErr,
				failure,
			)

			return nil
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
	require.Zero(
		t,
		runs,
	)
}

func TestFakePanicIsLogged(t *testing.T) {
	logger := &recordingLogger{}
	fake := &Fake{
		Logger: logger,
	}

	err := fake.Do(
		context.Background(),
		"root",
		func(context.Context) error {
			panic("boom")
		},
	)

	require.ErrorIs(
		t,
		err,
		txmanager.ErrPanic,
	)
	require.Equal(
		t,
		[]string{"panic occurred during transaction"},
		logger.errors,
	)
}

func TestFakeFlatFailureSwallowedTwoLevelsDeep(t *testing.T) {
	fake := New()
	failure := errors.New("deep failed")

	err := fake.Do(
		txmanager.Flat(context.Background()),
		"root",
		func(ctx context.Context) error {
			return fake.Do(
				ctx,
				"middle",
				func(ctx context.Context) error {
					require.ErrorIs(
						t,
						fake.Do(
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
		txmanager.ErrFlatNestedFailed,
	)
	require.ErrorIs(
		t,
		err,
		failure,
	)
}
