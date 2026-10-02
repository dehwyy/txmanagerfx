package gormtx

import (
	"time"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

const defaultTimeout = 1 * time.Minute

type Logger = txcore.Logger

type Opts struct {
	fx.In

	DB      *gorm.DB
	Logger  Logger        `optional:"true"`
	Timeout time.Duration `optional:"true"`
}

type TxManager struct {
	db      *gorm.DB
	logger  Logger
	timeout time.Duration
}

func New(opts Opts) (*TxManager, error) {
	logger := opts.Logger
	if logger == nil {
		logger = txcore.NopLogger()
	}

	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	return &TxManager{
		db:      opts.DB,
		logger:  logger,
		timeout: timeout,
	}, nil
}
