package txmanagertest

func (f *Fake) CallCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	return len(f.calls)
}
