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

	var existing *gorm.DB
	if conn, ok := txcore.Conn(ctx); ok {
		existing = conn.WithContext(ctx)
	}

	if existing != nil && applied.Attempts > 0 {
		return ErrRetryInNested
	}

	attempts := max(
		applied.Attempts,
		1,
	)

	backoff := applied.Backoff
	if backoff == nil {
		backoff = txcore.DefaultBackoff
	}

	var err error

	for attempt := 1; attempt <= attempts; attempt++ {
		err = tx.attempt(
			ctx,
			txName,
			existing,
			applied,
			fn,
		)
		if err == nil || attempt == attempts || !txcore.Retryable(err) {
			break
		}

		delay := backoff(attempt)

		tx.logger.Warn(
			"transaction retry",
			"tx_name",
			txName,
			"attempt",
			attempt,
			"backoff",
			delay,
			"error",
			err,
		)

		if waitErr := txcore.Wait(
			ctx,
			delay,
		); waitErr != nil {
			return err
		}
	}

	if err != nil && ctx.Err() != nil {
		tx.logger.Warn(
			"transaction context timeout/cancelled",
			"tx_name",
			txName,
			"error",
			ctx.Err(),
		)
	}

	return err
}

func (tx *TxManager) attempt(
	ctx context.Context,
	txName string,
	existing *gorm.DB,
	applied txcore.Options,
	fn func(ctx context.Context) error,
) error {
	if existing == nil {
		return tx.attemptRoot(
			ctx,
			txName,
			fn,
		)
	}

	if txcore.IsFlat(ctx) && !applied.Savepoint {
		return tx.runFlat(
			ctx,
			txName,
			fn,
		)
	}

	return tx.attemptSavepoint(
		ctx,
		txName,
		existing,
		fn,
	)
}
