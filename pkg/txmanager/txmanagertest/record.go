package txmanagertest

func (f *Fake) record(call Call) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls = append(f.calls, call)
}
