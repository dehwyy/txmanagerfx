package txmanager

import (
	"context"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
)

func Flat(ctx context.Context) context.Context {
	return txcore.Flat(ctx)
}
