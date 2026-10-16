package txcore

func OpenRoot(name string) *Scope {
	return &Scope{
		root: &root{
			name: name,
		},
		name: name,
	}
}
