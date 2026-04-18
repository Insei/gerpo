# Layers

A request flows top-down through four layers. Each layer has a narrow job and communicates with the next one through a small interface.

```
Repository[T]           (gerpo/*.go)
    │  user calls: GetFirst / GetList / Count / Insert / Update / Delete
    ▼
query/*Helper[T]        (query/*.go)
    │  per-request helpers that collect user intent
    ▼
query/linq              (query/linq/*.go)
    │  struct-based Where/Order/Exclude builders
    ▼
sqlstmt + sqlpart       (sqlstmt/*.go, sqlstmt/sqlpart/*.go)
    │  emit SQL text and bound arguments
    ▼
executor                (executor/*.go)
    │  run the statement, scan rows into the model, manage the cache
    ▼
executor/adapters/*     (executor/adapters/pgx5 | pgx4 | databasesql)
    │  driver-specific IO and placeholder rewriting
    ▼
database
```

## Public layer — `Repository[T]`

Source: `repository.go`, `builder.go`, `types.go`, `options.go`, `soft.go`, `column.go`.

Every public method follows the same recipe:

1. Obtain a pooled `stmt` object from `sqlstmt` (`NewGetFirst`, `NewInsert`, …).
2. `defer stmt.Release()` to return the object to its `sync.Pool`.
3. Apply the persistent query: `r.persistentQuery.Apply(stmt)`.
4. Apply the per-call `query.<X>Helper`: `q.Apply(stmt)`.
5. Hand the statement over to the executor.
6. Run `afterSelect` / `afterInsert` / `afterUpdate` as appropriate.
7. Pass errors through `errorTransformer`.

Every `func` in this layer is a thin orchestration step — no SQL, no reflection.

## `query/*Helper[T]`

Source: `query/*.go`.

One helper per operation: `GetFirstHelper`, `GetListHelper`, `CountHelper`, `InsertHelper`, `UpdateHelper`, `DeleteHelper`, and the special `PersistentHelper` for `WithQuery`.

They do not run queries themselves — they only collect user intent into structured objects from `query/linq` (WhereBuilder, OrderBuilder, ExcludeBuilder, PaginationBuilder). When `Apply` runs, those builders walk their internal op-slices and push work into `sqlpart` builders.

## `query/linq` — struct-based builders

Source: `query/linq/*.go`.

Prior to an internal refactor (see the `perf: replace closures…` commit), every operation was stored as a closure. Now each builder keeps a typed slice of `<kind>OpEntry` structs, and `Apply` is a `switch` dispatching on `kind`. This saved one allocation per condition and made the flow traceable in a debugger.

## `sqlstmt` + `sqlpart`

Source: `sqlstmt/*.go` and `sqlstmt/sqlpart/*.go`.

`sqlstmt` has one type per operation (`GetFirst`, `GetList`, `Count`, `Insert`, `Update`, `Delete`) that knows the SQL shape of its statement. `sqlpart` supplies reusable assemblers: `WhereBuilder`, `OrderBuilder`, `JoinBuilder`, `GroupBuilder`, `LimitOffsetBuilder`.

Every `sqlstmt` struct exposes `SQL() (string, []any, error)` — the final text and bound arguments. `GetFirst/GetList/Count` also have a `Release()` method that resets state and returns to a `sync.Pool`. See [SQL generation](sql-generation.md).

## `executor`

Source: `executor/executor.go`, `executor/types.go`, `executor/cache.go`.

The executor is the one place where IO happens. Responsibilities:

- Call the cache (`get`/`set`/`clean`) if a `cache.Storage` is wired.
- Delegate `ExecContext`/`QueryContext` to the adapter (or to a `Tx` when we're inside a transaction).
- Scan result rows via `stmt.Columns().GetModelPointers(model)`.

Caching, if any, is driven entirely by the `cache.Storage` interface — no assumptions about the backend.

## `executor/adapters/*`

Source: `executor/adapters/{pgx5,pgx4,databasesql,placeholder}/*.go`.

Three things live here:

- `NewPoolAdapter` / `NewAdapter` constructors producing a `DBAdapter`.
- `poolWrap` / `dbWrap` that translate the interface into driver-specific calls and rewrite placeholders.
- `txWrap` implementing `Tx` on top of the driver's own transaction type.

See [Adapter internals](adapters-internals.md).
