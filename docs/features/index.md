# Features

Reference of gerpo capabilities grouped by area.

!!! tip "See it all working together"
    [`examples/todo-api/`](https://github.com/Insei/gerpo/tree/main/examples/todo-api) is a runnable CRUD REST service that combines most of the pages below — column bindings, RETURNING, request-scope cache, transactions via `RunInTx`, domain error mapping — against a real PostgreSQL with goose migrations. `docker compose up --build` and the API boots on `:8080`. Pair it with the [Production-ready setup](../production-setup.md) page for the narrated walkthrough.

## Configuration

| Page | What's inside |
|---|---|
| [Repository builder](repository.md) | `New[T]()`, `DB`, `Table`, `Build`, repository lifecycle |
| [Columns](columns.md) | `Field`, `AsVirtual`, `OmitOnInsert`/`OmitOnUpdate`/`ReadOnly`, aliases, columns from other tables |
| [Persistent queries](persistent-queries.md) | `WithQuery`: conditions, JOINs, auto GROUP BY |
| [Soft delete](soft-delete.md) | Turning DELETE into UPDATE |
| [Virtual columns](virtual-columns.md) | Computed fields at the SELECT level |
| [Hooks](hooks.md) | Before/After for Insert/Update/Select |
| [Error transformer](error-transformer.md) | Mapping gerpo errors to domain errors |

## Operations

| Page | What's inside |
|---|---|
| [CRUD operations](crud.md) | `GetFirst`, `GetList`, `Count`, `Insert`, `Update`, `Delete` |
| [WHERE operators](where.md) | EQ, NotEQ, LT/LTE/GT/GTE, In/NotIn, Contains/StartsWith/EndsWith (+Fold variants), AND/OR/Group |
| [Ordering & pagination](order-pagination.md) | `OrderBy`, `Page`, `Size` |
| [Exclude & Only](exclude-only.md) | Narrowing columns in SELECT/INSERT/UPDATE |
| [Transactions](transactions.md) | `BeginTx`, `gerpo.WithTx(ctx, tx)`, `gerpo.RunInTx`, `Commit`, `Rollback`, `RollbackUnlessCommitted` |

## Infrastructure

| Page | What's inside |
|---|---|
| [Cache](cache.md) | `Cache` — cache scoped to a request context |
| [Tracing](tracing.md) | `WithTracer` hook — OpenTelemetry / Datadog / any tracer |
| [Adapters](adapters.md) | pgx v5, pgx v4, database/sql, and custom adapters |
