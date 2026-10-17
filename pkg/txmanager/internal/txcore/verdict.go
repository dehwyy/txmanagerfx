package txcore

import (
	"errors"
	"fmt"
)

func (s *Scope) Verdict(logger Logger, err error) error {
	if s == nil {
		return err
	}

	s.root.mu.Lock()
	failed := s.failed
	failedName := s.failedName
	failedErr := s.failedErr
	s.root.mu.Unlock()

	if !failed {
		return err
	}

	if err != nil && errors.Is(
		err,
		failedErr,
	) {
		return err
	}

	if s.savepoint {
		logger.Warn(
			"savepoint area rolled back after flat nested failure",
			"tx_name",
			s.root.name,
			"savepoint_tx_name",
			s.name,
			"nested_tx_name",
			failedName,
			"error",
			failedErr,
		)
	} else {
		logger.Warn(
			"flat nested transaction failed, root rolled back",
			"tx_name",
			s.root.name,
			"nested_tx_name",
			failedName,
			"error",
			failedErr,
		)
	}

	verdict := fmt.Errorf(
		"%w: %s: %w",
		ErrFlatNestedFailed,
		failedName,
		failedErr,
	)

	if err == nil {
		return verdict
	}

	return fmt.Errorf(
		"%w: %w",
		err,
		verdict,
	)
}
