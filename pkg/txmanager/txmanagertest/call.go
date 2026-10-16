package txmanagertest

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager"
)

func (f *Fake) call(
	ctx context.Context,
	txName string,
	fn func(ctx context.Context) error,
) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			f.logger().Error(
				"panic occurred during transaction",
				"panic",
				fmt.Sprint(recovered),
				"tx_name",
				txName,
				"stack",
				string(debug.Stack()),
			)

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
