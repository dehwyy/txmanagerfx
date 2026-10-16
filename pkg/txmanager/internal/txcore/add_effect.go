package txcore

import "context"

func AddEffect(ctx context.Context, effect func(context.Context) error) bool {
	found := lookup(ctx)
	if !found.active || found.scope == nil {
		return false
	}

	state := found.scope.root

	state.mu.Lock()
	defer state.mu.Unlock()

	if state.closed {
		return false
	}

	state.effects = append(
		state.effects,
		effect,
	)

	return true
}
