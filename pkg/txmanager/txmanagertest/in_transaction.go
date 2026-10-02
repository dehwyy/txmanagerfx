package txmanagertest

import (
	"context"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
)

func (f *Fake) InTransaction(ctx context.Context) bool {
	return txcore.Active(ctx)
}
