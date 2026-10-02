package txmanagertest

import (
	"context"
	"fmt"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager"
	"github.com/dehwyy/txmanagerfx/pkg/txmanager/internal/txcore"
)

func (f *Fake) Do(
	ctx context.Context,
	txName string,
	fn func(ctx context.Context) error,
	opts ...txmanager.Option,
) (err error) {
	nested := txcore.Active(ctx)

	f.record(Call{
		Name:   txName,
		Nested: nested,
	})

	if nested && txcore.Apply(opts).Attempts > 0 {
		return txmanager.ErrRetryInNested
	}

	if f.DoErr != nil {
		return f.DoErr
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%w in %s: %v", txmanager.ErrPanic, txName, r)
		}
	}()

	if err := fn(txcore.MarkActive(ctx)); err != nil {
		return err
	}

	return f.CommitErr
}
