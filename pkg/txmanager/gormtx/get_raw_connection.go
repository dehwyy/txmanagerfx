package gormtx

import (
	"context"

	"gorm.io/gorm"
)

func (tx *TxManager) GetRawConnection(ctx context.Context) *gorm.DB {
	return tx.db.WithContext(ctx)
}
