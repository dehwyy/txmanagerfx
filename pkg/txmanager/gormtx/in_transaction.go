package gormtx

import (
	"context"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
)

func (tx *TxManager) InTransaction(ctx context.Context) bool {
	return txcore.Active(ctx)
}
