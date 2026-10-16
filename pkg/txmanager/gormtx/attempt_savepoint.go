package gormtx

import (
	"context"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
	"gorm.io/gorm"
)

func (tx *TxManager) attemptSavepoint(
	ctx context.Context,
	txName string,
	existing *gorm.DB,
	fn func(ctx context.Context) error,
) error {
	parent := txcore.ScopeOf(ctx)

	var scope *txcore.Scope

	err := existing.Transaction(
		func(conn *gorm.DB) error {
			scope = parent.OpenSavepoint(txName)

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
	}

	return err
}
