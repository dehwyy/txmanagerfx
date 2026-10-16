package txmanagertest

import (
	"context"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager"
	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
)

func (f *Fake) Do(
	ctx context.Context,
	txName string,
	fn func(ctx context.Context) error,
	opts ...txmanager.Option,
) error {
	applied := txcore.Apply(opts)
	nested := txcore.Active(ctx)
	flat := txcore.IsFlat(ctx)

	f.record(Call{
		Name:      txName,
		Nested:    nested,
		Flat:      flat,
		Savepoint: applied.Savepoint,
	})

	if nested && applied.Attempts > 0 {
		return txmanager.ErrRetryInNested
	}

	if f.DoErr != nil {
		return f.DoErr
	}

	if !nested {
		return f.doRoot(
			ctx,
			txName,
			fn,
		)
	}

	if flat && !applied.Savepoint {
		return f.doFlat(
			ctx,
			txName,
			fn,
		)
	}

	return f.doSavepoint(
		ctx,
		txName,
		fn,
	)
}
