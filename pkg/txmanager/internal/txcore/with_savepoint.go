package txcore

func WithSavepoint() Option {
	return func(opts *Options) {
		opts.Savepoint = true
	}
}
