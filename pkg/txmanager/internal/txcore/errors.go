package txcore

import "errors"

var (
	ErrPanic         = errors.New("panic during transaction")
	ErrRetryInNested = errors.New("retry is allowed only on the root transaction")

	ErrFlatNestedFailed = errors.New("flat nested transaction failed")
)
