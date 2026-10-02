package txcore

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type stateErr struct {
	code string
}

func (e stateErr) Error() string {
	return e.code
}

func (e stateErr) SQLState() string {
	return e.code
}

func TestRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "deadlock", err: stateErr{code: "40P01"}, want: true},
		{name: "serialization", err: stateErr{code: "40001"}, want: true},
		{name: "unique violation", err: stateErr{code: "23505"}, want: false},
		{name: "plain error", err: errors.New("x"), want: false},
		{name: "wrapped deadlock", err: fmt.Errorf("wrap: %w", stateErr{code: "40P01"}), want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, Retryable(tt.err))
		})
	}
}
