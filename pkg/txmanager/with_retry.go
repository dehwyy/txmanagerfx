package txmanager

import "github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"

func WithRetry(attempts int) Option {
	return txcore.WithRetry(attempts)
}
