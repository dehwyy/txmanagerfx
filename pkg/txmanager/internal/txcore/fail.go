package txcore

func (s *Scope) Fail(name string, err error) {
	if s == nil {
		return
	}

	s.root.mu.Lock()
	defer s.root.mu.Unlock()

	if s.failed {
		return
	}

	s.failed = true
	s.failedName = name
	s.failedErr = err
}
