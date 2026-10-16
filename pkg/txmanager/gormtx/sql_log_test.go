package gormtx

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	sqlInsert     = "INSERT"
	sqlSelect     = "SELECT"
	sqlSavepoint  = "SAVEPOINT"
	sqlRollbackTo = "ROLLBACK TO"
	sqlOther      = "OTHER"
)

type sqlLog struct {
	mu         sync.Mutex
	statements []string
}

func (l *sqlLog) add(db *gorm.DB) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.statements = append(
		l.statements,
		sqlKind(db.Statement.SQL.String()),
	)
}

func (l *sqlLog) kinds() []string {
	l.mu.Lock()
	defer l.mu.Unlock()

	return append(
		[]string(nil),
		l.statements...,
	)
}

func (l *sqlLog) count(kind string) int {
	found := 0

	for _, statement := range l.kinds() {
		if statement == kind {
			found++
		}
	}

	return found
}

func sqlKind(statement string) string {
	switch {
	case strings.HasPrefix(
		statement,
		"ROLLBACK TO",
	):
		return sqlRollbackTo
	case strings.HasPrefix(
		statement,
		"SAVEPOINT",
	):
		return sqlSavepoint
	case strings.HasPrefix(
		statement,
		"INSERT",
	):
		return sqlInsert
	case strings.HasPrefix(
		statement,
		"SELECT",
	):
		return sqlSelect
	default:
		return sqlOther
	}
}

func recordSQL(t *testing.T, manager *TxManager) *sqlLog {
	t.Helper()

	log := &sqlLog{}
	callbacks := manager.db.Callback()

	require.NoError(
		t,
		callbacks.Create().After("gorm:create").Register(
			"txmanagerfx:test:create",
			log.add,
		),
	)
	require.NoError(
		t,
		callbacks.Query().After("gorm:query").Register(
			"txmanagerfx:test:query",
			log.add,
		),
	)
	require.NoError(
		t,
		callbacks.Raw().After("gorm:raw").Register(
			"txmanagerfx:test:raw",
			log.add,
		),
	)

	return log
}
