package txcore

import "context"

func WithScope(ctx context.Context, scope *Scope) context.Context {
	found := lookup(ctx)
	found.scope = scope

	return store(
		ctx,
		found,
	)
}
