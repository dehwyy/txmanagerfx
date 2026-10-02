package txcore

import (
	"math/rand/v2"
	"time"
)

const baseBackoff = 20 * time.Millisecond

func DefaultBackoff(attempt int) time.Duration {
	window := baseBackoff * time.Duration(attempt)

	return window + rand.N(window)
}
