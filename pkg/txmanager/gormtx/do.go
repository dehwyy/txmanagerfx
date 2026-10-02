package gormtx

import (
	"context"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
	"gorm.io/gorm"
)

func (tx *TxManager) Do(
	ctx context.Context,
	txName string,
	fn func(ctx context.Context) error,
	opts ...txcore.Option,
) error {
	applied := txcore.Apply(opts)

	root := tx.db
	nested := false
	if existing, ok := txcore.Conn(ctx); ok {
		root = existing.WithContext(ctx)
		nested = true
	}

	if nested && applied.Attempts > 0 {
		return ErrRetryInNested
	}

	attempts := max(applied.Attempts, 1)

	backoff := applied.Backoff
	if backoff == nil {
		backoff = txcore.DefaultBackoff
	}

	var err error

	for attempt := 1; attempt <= attempts; attempt++ {
		err = tx.attempt(
			ctx,
			txName,
			root,
			fn,
		)
		if err == nil || attempt == attempts || !txcore.Retryable(err) {
			break
		}

		delay := backoff(attempt)

		tx.logger.Warn(
			"transaction retry",
			"tx_name", txName,
			"attempt", attempt,
			"backoff", delay,
			"error", err,
		)

		if waitErr := txcore.Wait(ctx, delay); waitErr != nil {
			return err
		}
	}

	if err != nil && ctx.Err() != nil {
		tx.logger.Warn(
			"transaction context timeout/cancelled",
			"tx_name", txName,
			"error", ctx.Err(),
		)
	}

	return err
}

func (tx *TxManager) attempt(
	ctx context.Context,
	txName string,
	root *gorm.DB,
	fn func(ctx context.Context) error,
) error {
	return root.Transaction(
		func(conn *gorm.DB) error {
			return tx.run(
				ctx,
				txName,
				conn,
				fn,
			)
		},
	)
}
