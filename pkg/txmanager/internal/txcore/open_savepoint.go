package txcore

func (s *Scope) OpenSavepoint(name string) *Scope {
	if s == nil {
		return nil
	}

	s.root.mu.Lock()
	defer s.root.mu.Unlock()

	return &Scope{
		root:      s.root,
		name:      name,
		mark:      len(s.root.effects),
		savepoint: true,
	}
}
