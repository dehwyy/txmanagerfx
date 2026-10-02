package txmanager

import (
	"context"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
)

func Detach(ctx context.Context) context.Context {
	return txcore.Detach(ctx)
}
