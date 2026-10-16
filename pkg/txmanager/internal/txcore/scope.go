package txcore

import (
	"context"
	"sync"
)

type root struct {
	name    string
	mu      sync.Mutex
	effects []func(context.Context) error
	closed  bool
}

type Scope struct {
	root       *root
	name       string
	mark       int
	savepoint  bool
	failed     bool
	failedName string
	failedErr  error
}
