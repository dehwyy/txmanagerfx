package txmanager

import (
	"context"

	"gorm.io/gorm"
)

type TxManager interface {
	Do(
		ctx context.Context,
		txName string,
		fn func(ctx context.Context) error,
	) error

	GetConnection(
		ctx context.Context,
	) *gorm.DB

	GetRawConnection(
		ctx context.Context,
	) *gorm.DB
}
