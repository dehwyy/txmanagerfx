package txcore

const maxSubtransactions = 64

func (s *Scope) OpenSavepoint(name string, logger Logger) *Scope {
	if s == nil {
		return nil
	}

	s.root.mu.Lock()

	s.root.savepoints++
	count := s.root.savepoints

	exceeded := count > maxSubtransactions && !s.root.warned
	if exceeded {
		s.root.warned = true
	}

	scope := &Scope{
		root:      s.root,
		name:      name,
		mark:      len(s.root.effects),
		savepoint: true,
	}

	s.root.mu.Unlock()

	if exceeded {
		logger.Warn(
			"subtransaction count exceeded 64",
			"tx_name",
			s.root.name,
			"nested_tx_name",
			name,
			"count",
			count,
		)
	}

	return scope
}
