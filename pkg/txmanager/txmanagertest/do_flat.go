package txmanagertest

import (
	"context"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
)

func (f *Fake) doFlat(
	ctx context.Context,
	txName string,
	fn func(ctx context.Context) error,
) error {
	err := f.call(
		ctx,
		txName,
		fn,
	)
	if err != nil {
		txcore.ScopeOf(ctx).Fail(
			txName,
			err,
		)
	}

	return err
}
