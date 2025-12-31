package txmanager

import (
	"github.com/dehwyy/txmanagerfx/pkg/txmanager/gormtx"
	"go.uber.org/fx"
)

func NewGorm() fx.Option {
	return fx.Provide(
		func(opts gormtx.Opts) (*gormtx.TxManager, error) {
			return gormtx.New(opts)
		},
	)
}
