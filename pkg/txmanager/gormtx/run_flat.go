package gormtx

import (
	"context"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
)

func (tx *TxManager) runFlat(
	ctx context.Context,
	txName string,
	fn func(ctx context.Context) error,
) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = tx.panicked(
				txName,
				r,
			)
		}

		if err != nil {
			txcore.ScopeOf(ctx).Fail(
				txName,
				err,
			)
		}
	}()

	return fn(ctx)
}
