# Adapter internals

An adapter is a thin wrapper that turns `executor.DBAdapter` calls into driver-specific calls. All bundled adapters follow the same layout; differences come down to placeholder rewriting and transaction semantics.

## Anatomy of a wrapper

```
executor/adapters/<driver>/
    pool.go   — NewPoolAdapter / NewAdapter + ExecContext / QueryContext / BeginTx
    tx.go     — txWrap implementing executor.Tx
    rows.go   — rowsWrap adapting driver rows to types.Rows
    result.go — resultWrap adapting driver result to types.Result
```

The rest is boilerplate.

## Placeholder rewriting

gerpo's SQL uses `?`. Each adapter rewrites placeholders exactly once, right before handing the query to the driver:

```go
sql, err := placeholder.Dollar.ReplacePlaceholders(query)
```

`executor/adapters/placeholder/` provides two formats:

- `placeholder.Question` — no-op (input already `?`).
- `placeholder.Dollar` — scan-and-emit rewriter that turns `?` into `$1, $2, …`.

`databasesql.NewAdapter` defaults to `Question`. `pgx4` / `pgx5` always use `Dollar`.

## Rows wrapper

pgx returns `pgx.Rows`, `database/sql` returns `*sql.Rows`. Both shapes are close enough to the `types.Rows` interface, but they differ in `Scan` behavior (nullable types, text decoding). `rowsWrap` exists so gerpo can pretend both are identical.

## Result wrapper

`types.Result` exposes only `RowsAffected() (int64, error)`. Both pgx and `database/sql` return something richer, but gerpo only needs this one metric.

## Transaction wrapper

`txWrap` stores:

```go
type txWrap struct {
    commited                      bool
    rollbackUnlessCommittedNeeded bool
    tx                            <driver>.Tx // or *sql.Tx
}
```

- `Commit()` — calls driver commit, then sets `commited = true` on **success**.
- `Rollback()` — sets `rollbackUnlessCommittedNeeded = false`, then calls driver rollback.
- `RollbackUnlessCommitted()` — if `!commited && rollbackUnlessCommittedNeeded`, delegates to `Rollback()`; otherwise no-op. Designed to be safe as a `defer`.

All three methods use pointer receivers so the state mutations actually stick.

!!! warning "Historical bug"
    pgx v4 and v5 adapters originally used value receivers and also forgot to set `commited`. `RollbackUnlessCommitted()` after `Commit()` returned `tx is closed`. The integration test `TestTx_RollbackUnlessCommitted_AfterCommit` catches this; fixed in the `test: cover hooks, soft delete, …` commit.

## Writing your own

Walk through `executor/adapters/pgx5/` as a template. You will need:

- decide whether to rewrite placeholders (most non-PG drivers keep `?`; PG-derived drivers want `$N`);
- wrap the driver's `Rows` type in something satisfying `types.Rows`;
- wrap the driver's transaction type in `txWrap` following the rules above;
- return a `types.DBAdapter` implementation.

A good smoke test is `TestSmoke` in `tests/integration/` — `forEachAdapter` will pick up your new bundle once you add it to `allAdapters()`.
