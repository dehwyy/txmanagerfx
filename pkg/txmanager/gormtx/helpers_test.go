package gormtx

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type sqlStateError struct {
	code string
}

func (e *sqlStateError) Error() string {
	return fmt.Sprintf("sqlstate %s", e.code)
}

func (e *sqlStateError) SQLState() string {
	return e.code
}

type recordingLogger struct {
	mu       sync.Mutex
	warns    []string
	errorMsg []string
}

func (l *recordingLogger) Warn(msg string, _ ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.warns = append(l.warns, msg)
}

func (l *recordingLogger) Error(msg string, _ ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.errorMsg = append(l.errorMsg, msg)
}

func newFileManager(t *testing.T, opts Opts) *TxManager {
	t.Helper()

	dsn := fmt.Sprintf(
		"file:%s?_pragma=busy_timeout(5000)",
		filepath.Join(t.TempDir(), "tx.db"),
	)

	db, err := gorm.Open(
		sqlite.Open(dsn),
		&gorm.Config{},
	)
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)

	sqlDB.SetMaxOpenConns(4)

	t.Cleanup(
		func() {
			require.NoError(t, sqlDB.Close())
		},
	)

	require.NoError(t, db.AutoMigrate(&item{}))

	opts.DB = db

	manager, err := New(opts)
	require.NoError(t, err)

	return manager
}

func noBackoff(int) time.Duration {
	return 0
}
