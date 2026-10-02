package txmanagertest

import (
	"sync"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager"
	"gorm.io/gorm"
)

var _ txmanager.TxManager = (*Fake)(nil)

type Call struct {
	Name   string
	Nested bool
}

type Fake struct {
	DB *gorm.DB

	DoErr     error
	CommitErr error

	mu    sync.Mutex
	calls []Call
}

func New() *Fake {
	return &Fake{}
}
