package txmanager

import (
	"context"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
)

func AfterCommit(ctx context.Context, fn func(context.Context) error) bool {
	return txcore.AddEffect(
		ctx,
		fn,
	)
}
