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
		opts ...Option,
	) error

	GetConnection(
		ctx context.Context,
	) *gorm.DB

	GetRawConnection(
		ctx context.Context,
	) *gorm.DB

	InTransaction(
		ctx context.Context,
	) bool
}
