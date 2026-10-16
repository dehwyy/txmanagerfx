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
		if recovered := recover(); recovered != nil {
			err = tx.panicked(
				txName,
				recovered,
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
