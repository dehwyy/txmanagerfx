package txmanager

import (
	"context"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
)

func InTransaction(ctx context.Context) bool {
	return txcore.Active(ctx)
}
