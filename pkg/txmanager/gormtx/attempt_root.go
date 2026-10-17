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
			runErr := tx.run(
				txcore.WithScope(
					ctx,
					scope,
				),
				txName,
				conn,
				fn,
			)

			return scope.Verdict(
				tx.logger,
				runErr,
			)
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
