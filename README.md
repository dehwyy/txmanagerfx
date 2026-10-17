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

Free functions and options: `txmanager.Detach`, `txmanager.InTransaction`, `txmanager.AfterCommit`, `txmanager.Flat`, `txmanager.WithRetry`, `txmanager.WithBackoff`, `txmanager.WithSavepoint`.

Without `Flat` and `AfterCommit` the SQL is the same as in v0.1.0.

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

### AfterCommit

```go
manager.Do(ctx, "hold", func(ctx context.Context) error {
    if err := repo.Save(ctx, order); err != nil {
        return err
    }

    txmanager.AfterCommit(ctx, func(ctx context.Context) error {
        return timers.Schedule(ctx, order.ID)
    })

    return nil
})
```

`txmanager.AfterCommit(ctx, fn) bool` registers `fn` on the root transaction of `ctx`.

- Outside a transaction it returns `false` and does not call `fn`. This includes a `Detach`ed context and a context kept after its root `Do` returned. The caller decides whether to run the effect right away.
- Inside a transaction it returns `true`. Effects run in the root `Do` after a successful COMMIT, synchronously, before `Do` returns, in registration order, each at most once.
- The effect context is the context passed to the root `Do`: it carries no connection and `InTransaction` is false, so a `Do` inside an effect opens a new root transaction.
- An effect error is logged as `Warn("after-commit effect failed", tx_name, error)`, a panic as `Error("after-commit effect panicked", tx_name, panic, stack)`. The remaining effects still run and `Do` returns nil.
- Effects are dropped when the root `fn` returns an error or panics, when COMMIT fails, when the root is rolled back as rollback-only (see Flat) and on every retried attempt: only the attempt that commits runs its effects.
- When a savepoint area rolls back (a nested `Do` returned an error), effects registered inside it are dropped; effects registered before it stay.
- Registration is safe from several goroutines of one transaction.

An ambiguous COMMIT (the connection is lost after COMMIT was sent) returns an error, so no effects run although the data may be committed. An effect must never be the only path to its outcome: the service needs a recovery path for committed rows whose effect did not run (a reconciler, an outbox relay, a timer sweep).

### Flat

```go
err := manager.Do(txmanager.Flat(ctx), "apply-batch", func(ctx context.Context) error {
    for _, msg := range batch {
        if err := manager.Do(ctx, "apply-one", apply(msg)); err != nil {
            return err
        }
    }

    return nil
})
```

`txmanager.Flat(ctx)` marks the context: nested `Do` calls below it join the transaction without SAVEPOINT. The marker can be set before the root `Do` or inside it; the root `Do` itself does not change. `Detach` clears the marker.

A flat nested `Do`:

- runs `fn` on the same connection, without SAVEPOINT and without its own timeout (the root deadline applies);
- converts a panic to an error wrapping `ErrPanic`;
- on an error or a panic marks the current area rollback-only, even when the caller swallows the error. The area is the root or the nearest `WithSavepoint` nested `Do`. Without SAVEPOINT the database cannot undo part of a transaction, so committing after a swallowed failure would commit a partial transaction;
- rejects `WithRetry` with `ErrRetryInNested` without running `fn` and without marking.

When the root `fn` returns nil while its area is rollback-only, the root rolls back and returns an error wrapping both `txmanager.ErrFlatNestedFailed` and the original error (`errors.Is` matches both). It logs `Warn("flat nested transaction failed, root rolled back", tx_name, nested_tx_name, error)`: the calling code relied on SAVEPOINT semantics. When the root `fn` returns an error itself, that error is returned as usual, without the warning. A retryable original error (`40001`, `40P01`) makes a root with `WithRetry` try again.

In PostgreSQL a failed statement aborts the whole transaction: under Flat every later statement fails with `25P02` until the rollback.

### WithSavepoint

`txmanager.WithSavepoint()` forces a nested `Do` through SAVEPOINT even under Flat. Use it for writes whose failure the caller deliberately ignores.

- It opens a new area. On an error the database rolls back to the savepoint, the effects of the area are dropped, the error goes to the caller and the parent area is not marked.
- Flat nested calls inside it mark this area, not the root. When its `fn` returns nil while the area is rollback-only, the `Do` rolls back to the savepoint and returns `ErrFlatNestedFailed` wrapping the original error, and logs `Warn("savepoint area rolled back after flat nested failure", tx_name, savepoint_tx_name, nested_tx_name, error)`.
- On a root `Do` it is a no-op. Without Flat a nested `Do` already uses SAVEPOINT, so the option changes nothing.

### Subtransaction counter

Every SAVEPOINT issued by a nested `Do` is counted per root attempt. The count is cumulative, not the depth: RELEASE SAVEPOINT is never sent, and PostgreSQL keeps only 64 subtransaction ids per backend in shared memory before falling back to the subtransaction SLRU. The 65th SAVEPOINT logs `Warn("subtransaction count exceeded 64", tx_name, nested_tx_name, count)` once per attempt. The threshold is fixed. Flat nested calls are not counted.

### Logging

The library never writes to a global logger. Provide `gormtx.Logger` to receive events:

```go
type Logger interface {
    Warn(msg string, args ...any)
    Error(msg string, args ...any)
}
```

`*slog.Logger` satisfies it. Without a logger the events are dropped.

| Level | Message |
|---|---|
| Warn | `transaction retry` |
| Warn | `transaction context timeout/cancelled` |
| Warn | `after-commit effect failed` |
| Warn | `flat nested transaction failed, root rolled back` |
| Warn | `savepoint area rolled back after flat nested failure` |
| Warn | `subtransaction count exceeded 64` |
| Error | `panic occurred during transaction` |
| Error | `after-commit effect panicked` |

Only the root warning means that the whole root transaction was rolled back because of a Flat failure. Alert on the exact message `flat nested transaction failed, root rolled back`, not on the substring: the text of `ErrFlatNestedFailed` is `flat nested transaction failed` and also reaches the logs of callers that ignore the error of a `WithSavepoint` area.

## Testing

Package `txmanagertest`:

- `txmanagertest.New()` returns `*Fake`: runs `fn` with a context for which `InTransaction` is true, records `Calls()` (name and whether nested), supports `DoErr` (returned instead of running `fn`) and `CommitErr` (returned after a successful `fn`), converts panics to `ErrPanic`, rejects `WithRetry` in nested calls. `Fake.DB` optionally backs `GetConnection`.
- The Fake supports `AfterCommit`, `Flat`, `WithSavepoint`, rollback-only areas with `ErrFlatNestedFailed` and the subtransaction counter through the same internal code as `gormtx`. Effects run after a successful `fn` unless `DoErr` or `CommitErr` is set. `Call.Flat` is true when the context carried the Flat marker, `Call.Savepoint` when `WithSavepoint` was passed. The optional `Fake.Logger` receives the same events as `gormtx.Logger`.
- `NewMockTxManager(t)`: mockery mock with expecter.
