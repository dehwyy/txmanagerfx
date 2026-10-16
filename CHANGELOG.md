# Changelog

## v0.2.0

Added:
- `txmanager.AfterCommit(ctx, fn) bool`: effects registered on the root transaction run after a successful COMMIT, in order, before `Do` returns; dropped on rollback, COMMIT failure, retried attempts and savepoint rollback of their area.
- `txmanager.Flat(ctx)`: nested `Do` below the marker joins the transaction without SAVEPOINT; a failure marks the area rollback-only and the root rolls back with `ErrFlatNestedFailed`.
- `txmanager.WithSavepoint()`: nested `Do` keeps SAVEPOINT under Flat and isolates its area.
- `txmanager.ErrFlatNestedFailed` and `gormtx.ErrFlatNestedFailed`.
- Warn `subtransaction count exceeded 64`, once per root attempt.
- `txmanagertest`: Fake support for `AfterCommit`, `Flat`, `WithSavepoint`; `Call.Flat`, `Call.Savepoint`, optional `Fake.Logger`.
- SQL regression tests for the default nested SAVEPOINT path.

## v0.1.0

Breaking:
- `TxManager.Do` takes variadic `...Option`, and the interface has a new method `InTransaction(ctx) bool`. Callers compile unchanged; hand-written fakes and mocks of `TxManager` must add the new signature (use `txmanagertest.Fake` or `NewMockTxManager`).
- `github.com/rs/zerolog` is no longer used. Panic and timeout events are logged only through an optional `gormtx.Logger`.
- The context key is a private typed key instead of the string `tx-manager-gorm.connection`.

Added:
- `txmanager.Detach(ctx)`.
- `txmanager.InTransaction(ctx)` and `TxManager.InTransaction`.
- `txmanager.WithRetry`, `txmanager.WithBackoff`, retry on SQLSTATE `40P01` and `40001`, `ErrRetryInNested`.
- `gormtx.Opts.Logger` and `gormtx.Opts.Timeout` (both optional).
- Package `txmanagertest`: `Fake` and a mockery mock.
- Tests for nested SAVEPOINT, including three levels.

## v0.0.5

- Nested `Do` joins the outer transaction via SAVEPOINT.
- Panic inside `fn` returns an error and rolls back.
