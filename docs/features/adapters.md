# Adapters

gerpo never talks to a specific driver directly — it communicates through the `executor.Adapter` interface. Three implementations ship in the box.

## Bundled adapters

### pgx v5

```go
import (
    "github.com/insei/gerpo/executor/adapters/pgx5"
    "github.com/jackc/pgx/v5/pgxpool"
)

pool, _ := pgxpool.New(ctx, dsn)
adapter := pgx5.NewPoolAdapter(pool)
```

Placeholders: `$1, $2, …`.

### pgx v4

```go
import (
    "github.com/insei/gerpo/executor/adapters/pgx4"
    "github.com/jackc/pgx/v4/pgxpool"
)

pool, _ := pgxpool.Connect(ctx, dsn)
adapter := pgx4.NewPoolAdapter(pool)
```

Identical API, just a different pgx major.

### database/sql

A universal adapter for any `*sql.DB`. Defaults to `?` placeholders (MySQL-compatible). For PostgreSQL switch to `$1` explicitly:

```go
import (
    "database/sql"
    _ "github.com/jackc/pgx/v5/stdlib"

    "github.com/insei/gerpo/executor/adapters/databasesql"
    "github.com/insei/gerpo/executor/adapters/placeholder"
)

db, _ := sql.Open("pgx", dsn)
adapter := databasesql.NewAdapter(db, databasesql.WithPlaceholder(placeholder.Dollar))
```

## The `Adapter` interface

To write a custom adapter — implement three methods:

```go
type Adapter interface {
    ExecContext(ctx context.Context, query string, args ...any) (Result, error)
    QueryContext(ctx context.Context, query string, args ...any) (Rows, error)
    BeginTx(ctx context.Context) (Tx, error)
}
```

`Result`, `Rows`, `Tx` live in `executor/types`:

```go
type Rows interface {
    Next() bool
    Scan(dest ...any) error
    Close() error
}

type Result interface {
    RowsAffected() (int64, error)
}

type Tx interface {
    ExecQuery
    Commit() error
    Rollback() error
    RollbackUnlessCommitted() error
}
```

## Why write a custom adapter

- **Tracing** — wrap an existing adapter and add spans/logs around `ExecContext`/`QueryContext`.
- **A different driver** — ClickHouse, SQLite, MSSQL.
- **Mocks** — this is exactly how the mock benchmarks and some unit tests are wired (see `tests/mockdb_test.go`).

A small tracing wrapper:

```go
type tracingAdapter struct {
    inner executor.Adapter
    tr    trace.Tracer
}

func (a *tracingAdapter) QueryContext(ctx context.Context, q string, args ...any) (types.Rows, error) {
    ctx, span := a.tr.Start(ctx, "db.query")
    span.SetAttributes(attribute.String("db.statement", q))
    rows, err := a.inner.QueryContext(ctx, q, args...)
    if err != nil {
        span.RecordError(err)
    }
    span.End()
    return rows, err
}
// same idea for ExecContext and BeginTx
```

## Placeholder rewriting

Internally gerpo emits `?` placeholders. Each adapter decides whether to rewrite them:

- `pgx5` / `pgx4` — rewrite to `$1, $2, …` (`placeholder.Dollar`).
- `databasesql` — configurable via `WithPlaceholder(placeholder.Question | placeholder.Dollar)`.

If your driver accepts `?`, no rewriting is needed.
