package gormtx

import (
	"fmt"
	"runtime/debug"
)

func (tx *TxManager) panicked(txName string, recovered any) error {
	tx.logger.Error(
		"panic occurred during transaction",
		"panic",
		fmt.Sprint(recovered),
		"tx_name",
		txName,
		"stack",
		string(debug.Stack()),
	)

	return fmt.Errorf(
		"%w in %s: %v",
		ErrPanic,
		txName,
		recovered,
	)
}
