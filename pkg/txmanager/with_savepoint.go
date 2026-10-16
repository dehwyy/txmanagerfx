package txmanager

import "github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"

func WithSavepoint() Option {
	return txcore.WithSavepoint()
}
