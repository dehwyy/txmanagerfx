package txcore

import (
	"context"

	"gorm.io/gorm"
)

type ctxKey struct{}

type entry struct {
	conn   *gorm.DB
	active bool
}

func lookup(ctx context.Context) entry {
	v, ok := ctx.Value(ctxKey{}).(entry)
	if !ok {
		return entry{}
	}

	return v
}

func WithConn(ctx context.Context, conn *gorm.DB) context.Context {
	return context.WithValue(
		ctx,
		ctxKey{},
		entry{
			conn:   conn,
			active: true,
		},
	)
}

func MarkActive(ctx context.Context) context.Context {
	return context.WithValue(
		ctx,
		ctxKey{},
		entry{
			conn:   lookup(ctx).conn,
			active: true,
		},
	)
}

func Detach(ctx context.Context) context.Context {
	return context.WithValue(
		ctx,
		ctxKey{},
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
