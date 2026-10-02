package gormtx

import (
	"context"
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type item struct {
	ID   uint `gorm:"primaryKey"`
	Name string
}

func newTestManager(t *testing.T) *TxManager {
	t.Helper()

	db, err := gorm.Open(
		sqlite.Open("file::memory:"),
		&gorm.Config{},
	)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}

	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	if err := db.AutoMigrate(&item{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	manager, err := New(Opts{DB: db})
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	return manager
}

func insert(ctx context.Context, manager *TxManager, name string) error {
	return manager.GetConnection(ctx).Create(&item{Name: name}).Error
}

func names(t *testing.T, manager *TxManager) []string {
	t.Helper()

	var found []string
	err := manager.GetRawConnection(context.Background()).
		Model(&item{}).
		Order("id").
		Pluck("name", &found).Error
	if err != nil {
		t.Fatalf("select: %v", err)
	}

	return found
}

func assertNames(t *testing.T, manager *TxManager, want ...string) {
	t.Helper()

	got := names(t, manager)
	if len(got) != len(want) {
		t.Fatalf("want %v, got %v", want, got)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("want %v, got %v", want, got)
		}
	}
}

func TestNestedDoCommitsBoth(t *testing.T) {
	manager := newTestManager(t)

	err := manager.Do(context.Background(), "outer", func(ctx context.Context) error {
		if err := insert(ctx, manager, "outer"); err != nil {
			return err
		}

		return manager.Do(ctx, "inner", func(ctx context.Context) error {
			return insert(ctx, manager, "inner")
		})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertNames(t, manager, "outer", "inner")
}

func TestNestedDoInnerErrorRollsBackOnlyInner(t *testing.T) {
	manager := newTestManager(t)
	sentinel := errors.New("inner failed")

	err := manager.Do(context.Background(), "outer", func(ctx context.Context) error {
		if err := insert(ctx, manager, "outer-before"); err != nil {
			return err
		}

		innerErr := manager.Do(ctx, "inner", func(ctx context.Context) error {
			if err := insert(ctx, manager, "inner"); err != nil {
				return err
			}

			return sentinel
		})
		if !errors.Is(innerErr, sentinel) {
			t.Fatalf("expected sentinel, got %v", innerErr)
		}

		return insert(ctx, manager, "outer-after")
	})
	if err != nil {
		t.Fatalf("outer must stay usable and commit: %v", err)
	}

	assertNames(t, manager, "outer-before", "outer-after")
}

func TestNestedDoOuterErrorRollsBackEverything(t *testing.T) {
	manager := newTestManager(t)
	sentinel := errors.New("outer failed")

	err := manager.Do(context.Background(), "outer", func(ctx context.Context) error {
		if err := insert(ctx, manager, "outer"); err != nil {
			return err
		}

		if err := manager.Do(ctx, "inner", func(ctx context.Context) error {
			return insert(ctx, manager, "inner")
		}); err != nil {
			return err
		}

		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel, got %v", err)
	}

	assertNames(t, manager)
}

func TestNestedDoPanicRollsBackInnerOnly(t *testing.T) {
	manager := newTestManager(t)

	err := manager.Do(context.Background(), "outer", func(ctx context.Context) error {
		if err := insert(ctx, manager, "outer"); err != nil {
			return err
		}

		innerErr := manager.Do(ctx, "inner", func(ctx context.Context) error {
			if err := insert(ctx, manager, "inner"); err != nil {
				return err
			}

			panic("boom")
		})
		if !errors.Is(innerErr, ErrPanic) {
			t.Fatalf("expected ErrPanic, got %v", innerErr)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("outer must commit: %v", err)
	}

	assertNames(t, manager, "outer")
}

func TestTopLevelDoPanicRollsBack(t *testing.T) {
	manager := newTestManager(t)

	err := manager.Do(context.Background(), "outer", func(ctx context.Context) error {
		if err := insert(ctx, manager, "outer"); err != nil {
			return err
		}

		panic("boom")
	})
	if !errors.Is(err, ErrPanic) {
		t.Fatalf("expected ErrPanic, got %v", err)
	}

	assertNames(t, manager)
}

func TestNestedDoThreeLevelsSavepointChain(t *testing.T) {
	manager := newTestManager(t)
	sentinel := errors.New("level 3 failed")

	err := manager.Do(context.Background(), "l1", func(ctx context.Context) error {
		if err := insert(ctx, manager, "l1"); err != nil {
			return err
		}

		return manager.Do(ctx, "l2", func(ctx context.Context) error {
			if err := insert(ctx, manager, "l2"); err != nil {
				return err
			}

			l3Err := manager.Do(ctx, "l3", func(ctx context.Context) error {
				if err := insert(ctx, manager, "l3"); err != nil {
					return err
				}

				return sentinel
			})
			if !errors.Is(l3Err, sentinel) {
				t.Fatalf("expected sentinel, got %v", l3Err)
			}

			return insert(ctx, manager, "l2-after")
		})
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertNames(t, manager, "l1", "l2", "l2-after")
}
