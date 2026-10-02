# Changelog

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
