package gormtx

import (
	"context"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
	"gorm.io/gorm"
)

func (tx *TxManager) attemptRoot(
	ctx context.Context,
	txName string,
	fn func(ctx context.Context) error,
) error {
	scope := txcore.OpenRoot(txName)

	err := tx.db.Transaction(
		func(conn *gorm.DB) error {
			if err := tx.run(
				txcore.WithScope(
					ctx,
					scope,
				),
				txName,
				conn,
				fn,
			); err != nil {
				return err
			}

			return scope.Verdict(tx.logger)
		},
	)
	if err != nil {
		scope.Discard()

		return err
	}

	scope.RunEffects(
		ctx,
		tx.logger,
	)

	return nil
}
