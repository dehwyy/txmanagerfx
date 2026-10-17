package txmanagertest

import (
	"context"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
)

func (f *Fake) doSavepoint(
	ctx context.Context,
	txName string,
	fn func(ctx context.Context) error,
) error {
	scope := txcore.ScopeOf(ctx).OpenSavepoint(
		txName,
		f.logger(),
	)

	err := scope.Verdict(
		f.logger(),
		f.call(
			txcore.MarkActive(
				txcore.WithScope(
					ctx,
					scope,
				),
			),
			txName,
			fn,
		),
	)

	if err == nil {
		err = f.CommitErr
	}

	if err != nil {
		scope.Discard()
	}

	return err
}
