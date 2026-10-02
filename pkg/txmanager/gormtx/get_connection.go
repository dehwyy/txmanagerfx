package gormtx

import (
	"context"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
	"gorm.io/gorm"
)

func (tx *TxManager) GetConnection(ctx context.Context) *gorm.DB {
	conn, ok := txcore.Conn(ctx)
	if !ok {
		return tx.GetRawConnection(ctx)
	}

	return conn.WithContext(ctx)
}
