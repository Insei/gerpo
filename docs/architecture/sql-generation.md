# SQL generation

Every operation has a dedicated `sqlstmt` type; each type emits SQL by concatenating pieces produced by `sqlstmt/sqlpart` builders.

## Statement types

| Type | File | SQL shape |
|---|---|---|
| `GetFirst` | `sqlstmt/first.go` | `SELECT … FROM t [JOIN] [WHERE] [GROUP BY] [ORDER BY] LIMIT 1` |
| `GetList` | `sqlstmt/list.go` | `SELECT … FROM t [JOIN] [WHERE] [GROUP BY] [ORDER BY] [LIMIT … OFFSET …]` |
| `Count` | `sqlstmt/count.go` | `SELECT count(*) over() AS count FROM t [JOIN] [WHERE] [GROUP BY] LIMIT 1` |
| `Insert` | `sqlstmt/insert.go` | `INSERT INTO t (cols…) VALUES (?, …)` |
| `Update` | `sqlstmt/update.go` | `UPDATE t SET col = ?, … [WHERE]` |
| `Delete` | `sqlstmt/delete.go` | `DELETE FROM t [JOIN] [WHERE]` |

Each type implements `executor.Stmt` (or `CountStmt`) and exposes `SQL() (string, []any, error)`.

## sqlpart builders

`sqlstmt/sqlpart/` holds reusable fragment builders. Each of them keeps an internal buffer plus a `Reset(ctx)` method so the parent statement can recycle it across calls.

| Builder | Output |
|---|---|
| `WhereBuilder` | ` WHERE a = ? AND (b = ? OR c IS NULL)` |
| `JoinBuilder` | ` LEFT JOIN posts ON … INNER JOIN tags ON …` |
| `OrderBuilder` | ` ORDER BY created_at DESC, id ASC` |
| `GroupBuilder` | ` GROUP BY id, name` |
| `LimitOffsetBuilder` | ` LIMIT 20 OFFSET 40` |

Each returns an empty string when it has nothing to emit, so the parent can unconditionally concatenate without worrying about stray whitespace.

## Operator-to-SQL mapping

`sqlstmt/sqlpart/where.go` holds one factory per operator (`genEQFn`, `genLTFn`, `genINFn`, …). Factories are invoked once per column at repository build time, producing `func(ctx, value) (string, bool)` closures cached inside the column object. At query time a WHERE condition looks up the matching factory by operator name and appends the resulting SQL fragment with the argument (if any).

The LIKE family wraps parameters in `CAST(? AS text)` so PostgreSQL can infer the parameter type inside `CONCAT(…)`.

## Object pooling

`GetFirst`, `GetList`, and `Count` live in a `sync.Pool`. Lifecycle:

1. `NewGetFirst(ctx, table, cols)` → take from pool, call `reset(ctx, cols)`, return ready object.
2. Repository uses it.
3. `defer stmt.Release()` → zero mutable fields, return to pool.

`sqlselect` (the shared part of `GetFirst`/`GetList`/`Count`) owns an instance of each `sqlpart` builder. `sqlselect.reset` re-parents them to the new context and clears their internal buffers without giving up the backing memory — so a hot repository steadily reuses the same byte slices.

The pool can shrink under GC pressure; statement objects are designed to be cheap to allocate from scratch too, so there's no risk of starvation.

## Assembly flow (GetFirst example)

```
repository.GetFirst(ctx, qFns...)
 ├── NewGetFirst(ctx, table, columns)        // from pool
 ├── persistentQuery.Apply(stmt)               // WHERE, JOIN, GROUP BY injected
 ├── query.NewGetFirst(model).HandleFn(qFns...)
 │       .Apply(stmt)                          // per-request WHERE, ORDER, EXCLUDE
 ├── executor.GetOne(ctx, stmt)
 │       ├── stmt.SQL()
 │       ├── cache.get(ctx, sql, args...)     // optional
 │       ├── adapter.QueryContext(...)
 │       ├── rows.Scan(columns.GetModelPointers(m)...)
 │       └── cache.set(...)
 └── stmt.Release()                            // back to pool
```

The executor, the adapter, and the cache are interchangeable; the pipeline stays the same.
