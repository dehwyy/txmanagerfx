package txcore

import "context"

func IsFlat(ctx context.Context) bool {
	return lookup(ctx).flat
}
