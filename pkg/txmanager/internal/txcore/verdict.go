package txcore

import "fmt"

func (s *Scope) Verdict(logger Logger) error {
	if s == nil {
		return nil
	}

	s.root.mu.Lock()
	failed := s.failed
	failedName := s.failedName
	failedErr := s.failedErr
	s.root.mu.Unlock()

	if !failed {
		return nil
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

	return fmt.Errorf(
		"%w: %s: %w",
		ErrFlatNestedFailed,
		failedName,
		failedErr,
	)
}
