package txmanagertest

import (
	"context"
	"testing"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMockSatisfiesInterfaceAndExpectations(t *testing.T) {
	manager := NewMockTxManager(t)

	var _ txmanager.TxManager = manager

	manager.EXPECT().
		InTransaction(mock.Anything).
		Return(true).
		Once()

	require.True(t, manager.InTransaction(context.Background()))
}
