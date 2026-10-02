package txmanager

import "github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"

func WithBackoff(backoff Backoff) Option {
	return txcore.WithBackoff(backoff)
}
