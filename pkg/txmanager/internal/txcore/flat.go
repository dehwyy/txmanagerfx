package txcore

import "context"

func Flat(ctx context.Context) context.Context {
	found := lookup(ctx)
	found.flat = true

	return store(
		ctx,
		found,
	)
}
