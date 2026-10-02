package txcore

import "errors"

const (
	CodeDeadlockDetected    = "40P01"
	CodeSerializationFailed = "40001"
)

type sqlStater interface {
	SQLState() string
}

func Retryable(err error) bool {
	var stater sqlStater
	if !errors.As(err, &stater) {
		return false
	}

	switch stater.SQLState() {
	case CodeDeadlockDetected, CodeSerializationFailed:
		return true
	default:
		return false
	}
}
