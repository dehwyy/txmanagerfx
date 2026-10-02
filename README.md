# txmanagerfx

Transaction manager for GORM with fx wiring.

```go
fx.New(
    txmanager.NewGorm(),
)
```

Requires `*gorm.DB` in the graph. Optional: `gormtx.Logger` and `time.Duration` (transaction timeout, default 1m).

## API

```go
type TxManager interface {
    Do(ctx, txName, fn, opts ...Option) error
    GetConnection(ctx) *gorm.DB
    GetRawConnection(ctx) *gorm.DB
    InTransaction(ctx) bool
}
```

- `Do` opens a root transaction. A nested `Do` on a context that already carries a transaction joins it through SAVEPOINT: an inner error rolls back only the inner part, the outer transaction stays usable.
- A panic inside `fn` becomes an error wrapping `txmanager.ErrPanic` and rolls back.
- `GetConnection` returns the transaction from ctx, otherwise the raw connection. `GetRawConnection` is always non-transactional.
- `InTransaction(ctx)` is also available as `txmanager.InTransaction(ctx)`.

### Detach

`txmanager.Detach(ctx)` returns a context that no longer carries the transaction. `GetConnection` on it returns the raw connection, and `Do` opens a new root transaction. Use it for writes that must survive an outer rollback (audit rows, outbox failure records). Deadlines and values of the parent context are kept.

### Retry

```go
err := manager.Do(
    ctx,
    "move-funds",
    fn,
    txmanager.WithRetry(3),
    txmanager.WithBackoff(func(attempt int) time.Duration { ... }),
)
```

`WithRetry(n)` is the total number of attempts. Retried errors are those exposing `SQLState()` equal to `40P01` (deadlock) or `40001` (serialization failure), including wrapped ones; `*pgconn.PgError` qualifies. `fn` must be safe to run again. Default backoff is `20ms * attempt` plus equal jitter. Panics and other errors are not retried. `WithRetry` on a nested `Do` returns `txmanager.ErrRetryInNested` without running `fn`; on a `Detach`ed context it is allowed because that `Do` is a root.

### Logging

The library never writes to a global logger. Provide `gormtx.Logger` to receive events:

```go
type Logger interface {
    Warn(msg string, args ...any)
    Error(msg string, args ...any)
}
```

`*slog.Logger` satisfies it. Without a logger the events are dropped.

## Testing

Package `txmanagertest`:

- `txmanagertest.New()` returns `*Fake`: runs `fn` with a context for which `InTransaction` is true, records `Calls()` (name and whether nested), supports `DoErr` (returned instead of running `fn`) and `CommitErr` (returned after a successful `fn`), converts panics to `ErrPanic`, rejects `WithRetry` in nested calls. `Fake.DB` optionally backs `GetConnection`.
- `NewMockTxManager(t)`: mockery mock with expecter.
