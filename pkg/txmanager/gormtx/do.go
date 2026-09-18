package gormtx

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func (tx *TxManager) Do(
	ctx context.Context,
	txName string,
	fn func(ctx context.Context) error,
) error {
	err := tx.db.Transaction(
		func(conn *gorm.DB) error {
			return tx.run(
				ctx,
				txName,
				conn,
				fn,
			)
		},
	)

	if err != nil {
		if ctx.Err() != nil {
			log.Warn().
				Err(ctx.Err()).
				Msg("Transaction context timeout/cancelled")
		}
	}

	return err
}

func (tx *TxManager) run(
	ctx context.Context,
	txName string,
	conn *gorm.DB,
	fn func(ctx context.Context) error,
) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error().
				Any("panic", r).
				Str("tx_name", txName).
				Str("stack", string(debug.Stack())).
				Msg("Panic occurred during transaction")

			err = fmt.Errorf("%w in %s: %v", ErrPanic, txName, r)
		}
	}()

	ctx, cancel := context.WithTimeout(
		context.WithValue(ctx, contextKey, conn),
		defaultTimeout,
	)
	defer cancel()

	return fn(ctx)
}
