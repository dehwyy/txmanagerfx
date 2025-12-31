package gormtx

import (
	"time"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

const (
	defaultTimeout = 1 * time.Minute
	contextKey     = "tx-manager-gorm.connection"
)

type Opts struct {
	fx.In

	DB *gorm.DB
}

type TxManager struct {
	db *gorm.DB
}

func New(opts Opts) (*TxManager, error) {
	return &TxManager{
		db: opts.DB,
	}, nil
}
