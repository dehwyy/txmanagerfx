package gormtx

import (
	"context"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
	"gorm.io/gorm"
)

func (tx *TxManager) run(
	ctx context.Context,
	txName string,
	conn *gorm.DB,
	fn func(ctx context.Context) error,
) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = tx.panicked(
				txName,
				r,
			)
		}
	}()

	ctx, cancel := context.WithTimeout(
		txcore.WithConn(
			ctx,
			conn,
		),
		tx.timeout,
	)
	defer cancel()

	return fn(ctx)
}
