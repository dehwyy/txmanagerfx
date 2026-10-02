package gormtx

import (
	"context"
	"fmt"
	"runtime/debug"

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
			tx.logger.Error(
				"panic occurred during transaction",
				"panic", fmt.Sprint(r),
				"tx_name", txName,
				"stack", string(debug.Stack()),
			)

			err = fmt.Errorf("%w in %s: %v", ErrPanic, txName, r)
		}
	}()

	ctx, cancel := context.WithTimeout(
		txcore.WithConn(ctx, conn),
		tx.timeout,
	)
	defer cancel()

	return fn(ctx)
}
