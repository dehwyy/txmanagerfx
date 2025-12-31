package txmanager

import (
	"github.com/dehwyy/txmanagerfx/pkg/txmanager/gormtx"
	"go.uber.org/fx"
)

func NewGorm(opts gormtx.Opts) fx.Option {
	return fx.Provide(func() (*gormtx.TxManager, error) {
		return gormtx.New(opts)
	})
}
