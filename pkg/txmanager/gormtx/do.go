package gormtx

import (
	"context"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func (tx *TxManager) Do(
	ctx context.Context,
	txName string,
	fn func(ctx context.Context) error,
) error {
	err := tx.db.Transaction(
		func(tx *gorm.DB) error {
			defer func() {
				if r := recover(); r != nil {
					log.Error().
						Any("panic", r).
						Msg("Panic occurred during transaction")
				}
			}()

			ctx, cancel := context.WithTimeout(
				context.WithValue(ctx, contextKey, tx),
				defaultTimeout,
			)
			defer cancel()

			return fn(ctx)
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
