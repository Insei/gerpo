# Features

Reference of gerpo capabilities grouped by area.

## Configuration

| Page | What's inside |
|---|---|
| [Repository builder](repository.md) | `NewBuilder[T]()`, `DB`, `Table`, `Build`, repository lifecycle |
| [Columns](columns.md) | `AsColumn`, `AsVirtual`, insert/update protection, aliases, columns from other tables |
| [Persistent queries](persistent-queries.md) | `WithQuery`: conditions, JOINs, GROUP BY applied to every request |
| [Soft delete](soft-delete.md) | Turning DELETE into UPDATE |
| [Virtual columns](virtual-columns.md) | Computed fields at the SELECT level |
| [Hooks](hooks.md) | Before/After for Insert/Update/Select |
| [Error transformer](error-transformer.md) | Mapping gerpo errors to domain errors |

## Operations

| Page | What's inside |
|---|---|
| [CRUD operations](crud.md) | `GetFirst`, `GetList`, `Count`, `Insert`, `Update`, `Delete` |
| [WHERE operators](where.md) | EQ, NEQ, LT/LTE/GT/GTE, IN/NIN, CT/BW/EW + IC, AND/OR/Group |
| [Ordering & pagination](order-pagination.md) | `OrderBy`, `Page`, `Size` |
| [Exclude & Only](exclude-only.md) | Narrowing columns in SELECT/INSERT/UPDATE |
| [Transactions](transactions.md) | `BeginTx`, `repo.Tx`, `Commit`, `Rollback`, `RollbackUnlessCommitted` |

## Infrastructure

| Page | What's inside |
|---|---|
| [Cache](cache.md) | `CtxCache` — cache scoped to a request context |
| [Adapters](adapters.md) | pgx v5, pgx v4, database/sql, and custom adapters |
