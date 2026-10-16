package txmanagertest

import (
	"context"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
)

func (f *Fake) doRoot(
	ctx context.Context,
	txName string,
	fn func(ctx context.Context) error,
) error {
	scope := txcore.OpenRoot(txName)

	err := f.call(
		txcore.MarkActive(
			txcore.WithScope(
				ctx,
				scope,
			),
		),
		txName,
		fn,
	)
	if err == nil {
		err = scope.Verdict(f.logger())
	}

	if err == nil {
		err = f.CommitErr
	}

	if err != nil {
		scope.Discard()

		return err
	}

	scope.RunEffects(
		ctx,
		f.logger(),
	)

	return nil
}
