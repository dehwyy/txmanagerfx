package gormtx

import (
	"context"

	"gorm.io/gorm"
)

func (tx *TxManager) GetConnection(ctx context.Context) *gorm.DB {
	v, ok := ctx.Value(contextKey).(*gorm.DB)
	if !ok {
		return tx.GetRawConnection(ctx)
	}

	return v.WithContext(ctx)
}
