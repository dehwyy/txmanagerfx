package txmanagertest

import (
	"context"
	"fmt"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager"
)

func (f *Fake) call(
	ctx context.Context,
	txName string,
	fn func(ctx context.Context) error,
) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf(
				"%w in %s: %v",
				txmanager.ErrPanic,
				txName,
				recovered,
			)
		}
	}()

	return fn(ctx)
}
