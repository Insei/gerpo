# Adapter internals

An adapter turns `executor.Adapter` calls into underlying-driver calls. The placeholder rewrite, the transaction state machine and the `RollbackUnlessCommitted` semantics live once in the unexported `executor/adapters/internal` package; every bundled adapter (pgx v5, pgx v4, database/sql) only contributes a tiny `Driver` plus result/rows wrappers.

Terminology recap:

- **driver** — the SQL library (`pgx/v5`, `pgx/v4`, `database/sql`), or the internal `internal.Driver` interface that delegates to one.
- **adapter** — gerpo's public wrapper over a driver, implementing `executor.Adapter`.

## Anatomy of an adapter package

```
executor/adapters/<adapter>/
    pool.go    — internal.Driver / internal.TxDriver implementations + the public NewPoolAdapter / NewAdapter
    rows.go    — rowsWrap adapting driver rows to types.Rows (only when the driver's Rows
                 type doesn't already satisfy the interface)
    result.go  — resultWrap adapting driver result to types.Result (same caveat)
```

`databasesql` is the smallest of the three: `*sql.Rows` and `sql.Result` already satisfy `types.Rows` / `types.Result`, so no wrapper types are needed. pgx returns its own `pgx.Rows` / `pgconn.CommandTag`, which require thin wrappers.

## The shared base — `internal.Adapter`

`internal.New(driver Driver, p placeholder.PlaceholderFormat) extypes.Adapter` returns the public adapter. It owns:

- placeholder rewrite for every `ExecContext` / `QueryContext`;
- creation of a `transaction` wrapping the driver's `TxDriver`;
- the transaction state machine (`committed`, `rollbackUnlessCommittedNeeded`).

Drivers never reimplement that logic.

## The two driver interfaces

```go
type Driver interface {
    Exec(ctx context.Context, sql string, args ...any) (extypes.Result, error)
    Query(ctx context.Context, sql string, args ...any) (extypes.Rows, error)
    BeginTx(ctx context.Context) (TxDriver, error)
}

type TxDriver interface {
    Exec(ctx context.Context, sql string, args ...any) (extypes.Result, error)
    Query(ctx context.Context, sql string, args ...any) (extypes.Rows, error)
    Commit() error
    Rollback() error
}
```

A driver implements both with a few lines of delegation. `Commit` / `Rollback` are context-less because pgx insists on its own background context for these calls.

## Placeholder rewriting

gerpo emits `?` placeholders. The shared adapter rewrites them exactly once before delegating to the driver:

```go
sql, err := a.placeholder.ReplacePlaceholders(query)
```

`executor/adapters/placeholder/` provides two formats:

- `placeholder.Question` — no-op (input stays as `?`).
- `placeholder.Dollar` — scan-and-emit rewriter that turns `?` into `$1, $2, …`.

`databasesql.NewAdapter` defaults to `Question`; pass `WithPlaceholder(placeholder.Dollar)` for PostgreSQL. `pgx4` / `pgx5` always pin `Dollar`.

## Transaction state machine

```go
type transaction struct {
    inner                         TxDriver
    placeholder                   placeholder.PlaceholderFormat
    committed                     bool
    rollbackUnlessCommittedNeeded bool
}
```

- `Commit()` — calls `inner.Commit()`, then sets `committed = true` only on success.
- `Rollback()` — clears `rollbackUnlessCommittedNeeded`, then calls `inner.Rollback()`.
- `RollbackUnlessCommitted()` — if `!committed && rollbackUnlessCommittedNeeded`, delegates to `Rollback()`; otherwise no-op. Designed to be safe as a `defer`.

All three are pointer-receiver methods on the shared type, so state mutations actually persist (pgx wrappers historically used value receivers and lost the flag — fixed in `chore: fix "commited" typo in tx wrappers`).

## Rows / Result wrappers

`types.Rows` requires `Next()`, `Scan(dest ...any) error`, `Close() error`. `*sql.Rows` already matches this shape; pgx returns its own type with `Close()` returning nothing, so `rowsWrap` adapts it.

`types.Result` requires only `RowsAffected() (int64, error)`. `sql.Result` matches; pgx returns `pgconn.CommandTag` whose `RowsAffected()` returns just `int64`, so `resultWrap` adds the trailing `nil` error.

## Writing your own adapter

1. Implement `internal.Driver` (three methods) and `internal.TxDriver` (four methods) on top of the SQL driver you're wrapping.
2. Pick a placeholder format. Most non-PostgreSQL drivers keep `?` (`placeholder.Question`).
3. Wrap your driver's `Rows`/`Result` types only if their methods don't already satisfy the interfaces in `executor/types`.
4. Return `internal.New(yourDriver, yourPlaceholder)` from the public constructor.

A good smoke test is `TestSmoke` in `tests/integration/` — `forEachAdapter` will pick up your new bundle as soon as you add it to `allAdapters()`.

For unit-level coverage of the shared logic see `executor/adapters/internal/base_test.go` — it drives the adapter with a fake driver and exercises every transaction-lifecycle path.
