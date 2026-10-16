package txcore

import (
	"context"

	"gorm.io/gorm"
)

type ctxKey struct{}

type entry struct {
	conn   *gorm.DB
	active bool
	scope  *Scope
}

func lookup(ctx context.Context) entry {
	v, ok := ctx.Value(ctxKey{}).(entry)
	if !ok {
		return entry{}
	}

	return v
}

func store(ctx context.Context, found entry) context.Context {
	return context.WithValue(
		ctx,
		ctxKey{},
		found,
	)
}

func WithConn(ctx context.Context, conn *gorm.DB) context.Context {
	found := lookup(ctx)
	found.conn = conn
	found.active = true

	return store(
		ctx,
		found,
	)
}

func MarkActive(ctx context.Context) context.Context {
	found := lookup(ctx)
	found.active = true

	return store(
		ctx,
		found,
	)
}

func Detach(ctx context.Context) context.Context {
	return store(
		ctx,
		entry{},
	)
}

func Conn(ctx context.Context) (*gorm.DB, bool) {
	found := lookup(ctx)
	if !found.active || found.conn == nil {
		return nil, false
	}

	return found.conn, true
}

func Active(ctx context.Context) bool {
	return lookup(ctx).active
}
