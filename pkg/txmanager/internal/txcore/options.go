package txcore

import "time"

type Backoff func(attempt int) time.Duration

type Options struct {
	Attempts int
	Backoff  Backoff
}

type Option func(*Options)

func Apply(opts []Option) Options {
	var applied Options

	for _, opt := range opts {
		opt(&applied)
	}

	return applied
}

func WithRetry(attempts int) Option {
	return func(opts *Options) {
		opts.Attempts = attempts
	}
}

func WithBackoff(backoff Backoff) Option {
	return func(opts *Options) {
		opts.Backoff = backoff
	}
}
