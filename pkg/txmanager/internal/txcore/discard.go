package txcore

func (s *Scope) Discard() {
	if s == nil {
		return
	}

	s.root.mu.Lock()
	defer s.root.mu.Unlock()

	if s.mark < len(s.root.effects) {
		clear(s.root.effects[s.mark:])
		s.root.effects = s.root.effects[:s.mark]
	}

	if !s.savepoint {
		s.root.closed = true
	}
}
