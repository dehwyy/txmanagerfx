package txcore

import (
	"context"
	"fmt"
	"runtime/debug"
)

func (s *Scope) RunEffects(ctx context.Context, logger Logger) {
	if s == nil {
		return
	}

	s.root.mu.Lock()
	effects := s.root.effects
	s.root.effects = nil
	s.root.closed = true
	s.root.mu.Unlock()

	for _, effect := range effects {
		runEffect(
			ctx,
			logger,
			s.root.name,
			effect,
		)
	}
}

func runEffect(
	ctx context.Context,
	logger Logger,
	txName string,
	effect func(context.Context) error,
) {
	defer func() {
		if r := recover(); r != nil {
			logger.Error(
				"after-commit effect panicked",
				"tx_name",
				txName,
				"panic",
				fmt.Sprint(r),
				"stack",
				string(debug.Stack()),
			)
		}
	}()

	if err := effect(ctx); err != nil {
		logger.Warn(
			"after-commit effect failed",
			"tx_name",
			txName,
			"error",
			err,
		)
	}
}
