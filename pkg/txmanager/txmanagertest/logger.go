package txmanagertest

import (
	"github.com/dehwyy/txmanagerfx/pkg/txmanager"
	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
)

func (f *Fake) logger() txmanager.Logger {
	if f.Logger == nil {
		return txcore.NopLogger()
	}

	return f.Logger
}
