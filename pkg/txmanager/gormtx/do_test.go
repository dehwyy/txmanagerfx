package gormtx

import (
	"context"
	"errors"
	"testing"
)

func TestRunReturnsErrorWhenCallbackPanics(t *testing.T) {
	manager := &TxManager{}

	err := manager.run(
		context.Background(),
		"panicking",
		nil,
		func(ctx context.Context) error {
			panic("boom")
		},
	)

	if err == nil {
		t.Fatal("expected a non-nil error so the transaction rolls back, got nil")
	}

	if !errors.Is(err, ErrPanic) {
		t.Fatalf("expected the error to wrap ErrPanic, got %v", err)
	}
}

func TestRunReturnsErrorWhenCallbackPanicsWithError(t *testing.T) {
	manager := &TxManager{}

	err := manager.run(
		context.Background(),
		"panicking",
		nil,
		func(ctx context.Context) error {
			panic(errors.New("boom"))
		},
	)

	if !errors.Is(err, ErrPanic) {
		t.Fatalf("expected the error to wrap ErrPanic, got %v", err)
	}
}

func TestRunPassesThroughSuccess(t *testing.T) {
	manager := &TxManager{}

	err := manager.run(
		context.Background(),
		"ok",
		nil,
		func(ctx context.Context) error {
			return nil
		},
	)

	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestRunPassesThroughCallbackError(t *testing.T) {
	manager := &TxManager{}
	sentinel := errors.New("callback failed")

	err := manager.run(
		context.Background(),
		"failing",
		nil,
		func(ctx context.Context) error {
			return sentinel
		},
	)

	if !errors.Is(err, sentinel) {
		t.Fatalf("expected the callback error unchanged, got %v", err)
	}

	if errors.Is(err, ErrPanic) {
		t.Fatal("a plain callback error must not be reported as a panic")
	}
}
