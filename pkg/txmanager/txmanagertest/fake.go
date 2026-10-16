package txmanagertest

import (
	"sync"

	"github.com/dehwyy/txmanagerfx/pkg/txmanager"
	"gorm.io/gorm"
)

var _ txmanager.TxManager = (*Fake)(nil)

type Call struct {
	Name      string
	Nested    bool
	Flat      bool
	Savepoint bool
}

type Fake struct {
	DB     *gorm.DB
	Logger txmanager.Logger

	DoErr     error
	CommitErr error

	mu    sync.Mutex
	calls []Call
}

func New() *Fake {
	return &Fake{}
}
