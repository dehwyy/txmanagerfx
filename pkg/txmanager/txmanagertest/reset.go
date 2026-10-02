package txmanagertest

func (f *Fake) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls = nil
}
