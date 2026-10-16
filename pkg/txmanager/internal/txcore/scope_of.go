package txcore

import "context"

func ScopeOf(ctx context.Context) *Scope {
	found := lookup(ctx)
	if !found.active {
		return nil
	}

	return found.scope
}
