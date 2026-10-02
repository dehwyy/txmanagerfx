package txcore

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDefaultBackoffWindow(t *testing.T) {
	tests := []struct {
		attempt int
		min     time.Duration
		max     time.Duration
	}{
		{attempt: 1, min: 20 * time.Millisecond, max: 40 * time.Millisecond},
		{attempt: 2, min: 40 * time.Millisecond, max: 80 * time.Millisecond},
		{attempt: 3, min: 60 * time.Millisecond, max: 120 * time.Millisecond},
	}

	for _, tt := range tests {
		for range 50 {
			got := DefaultBackoff(tt.attempt)
			require.GreaterOrEqual(t, got, tt.min)
			require.Less(t, got, tt.max)
		}
	}
}
