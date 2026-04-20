# GERPO

[![codecov](https://codecov.io/gh/Insei/gerpo/graph/badge.svg?token=LGY9O9OJF5)](https://codecov.io/gh/Insei/gerpo)
[![build](https://github.com/Insei/gerpo/actions/workflows/go.yml/badge.svg)](https://github.com/Insei/gerpo/actions/workflows/go.yml)
[![Goreport](https://goreportcard.com/badge/github.com/insei/gerpo)](https://goreportcard.com/report/github.com/insei/gerpo)
[![GoDoc](https://godoc.org/github.com/insei/gerpo?status.svg)](https://godoc.org/github.com/insei/gerpo)
[![Docs](https://img.shields.io/badge/docs-insei.github.io%2Fgerpo-blue)](https://insei.github.io/gerpo/)

**GERPO** (Golang + Repository) is a generic repository pattern for Go with pluggable adapters and a tiny footprint. It is **not an ORM** — no migrations, no relations, no struct tags. All SQL behavior is declared once in the repository configuration; columns are bound to struct fields through pointers.

> **Database support.** gerpo currently targets **PostgreSQL** (and PG-compatible databases such as CockroachDB). The SQL fragments gerpo emits — placeholder format, LIKE type-casts, `RETURNING`, window functions — assume PostgreSQL. MySQL, MS SQL Server, and pre-3.35 SQLite are **not supported** today. See [`TODO.md`](TODO.md) for the multi-dialect backlog.

> 📚 Full documentation: **[insei.github.io/gerpo](https://insei.github.io/gerpo/)** · [Why gerpo?](https://insei.github.io/gerpo/why-gerpo/) (vs GORM / ent / bun / sqlc / sqlx) · API reference: **[pkg.go.dev/github.com/insei/gerpo](https://pkg.go.dev/github.com/insei/gerpo)**

## Install

```bash
go get github.com/insei/gerpo@latest
```

Minimum Go version: **1.24**.

## Quick start

```go
type User struct {
    ID        uuid.UUID
    Name      string
    Email     *string
    Age       int
    CreatedAt time.Time
}

repo, err := gerpo.New[User]().
    Adapter(pgx5.NewPoolAdapter(pool)).
    Table("users").
    Columns(func(m *User, c *gerpo.ColumnBuilder[User]) {
        c.Field(&m.ID).OmitOnUpdate()
        c.Field(&m.Name)
        c.Field(&m.Email)
        c.Field(&m.Age)
        c.Field(&m.CreatedAt).OmitOnUpdate()
    }).
    Build()

users, _ := repo.GetList(ctx, func(m *User, h query.GetListHelper[User]) {
    h.Where().Field(&m.Age).GTE(18)
    h.OrderBy().Field(&m.CreatedAt).DESC()
    h.Page(1).Size(20)
})
```

Full runnable samples live in [`examples/`](examples/) and in the [integration tests](tests/integration/).

## Features

| Area | Highlights | Docs |
|---|---|---|
| Repository | Type-safe builder, thread-safe, `sync.Pool` backed statements | [Repository builder](https://insei.github.io/gerpo/features/repository/) |
| Columns | `AsColumn` / `AsVirtual`, insert/update protection, aliases | [Columns](https://insei.github.io/gerpo/features/columns/), [Virtual columns](https://insei.github.io/gerpo/features/virtual-columns/) |
| Queries | 14 WHERE operators + IC variants, AND/OR/Group, ordering, pagination | [WHERE operators](https://insei.github.io/gerpo/features/where/), [Ordering & pagination](https://insei.github.io/gerpo/features/order-pagination/) |
| Operations | GetFirst / GetList / Count / Insert / InsertMany / Update / Delete with `Only` / `Exclude` | [CRUD operations](https://insei.github.io/gerpo/features/crud/), [Exclude & Only](https://insei.github.io/gerpo/features/exclude-only/) |
| Persistent queries | Always-on WHERE, JOIN, GROUP BY via `WithQuery` | [Persistent queries](https://insei.github.io/gerpo/features/persistent-queries/) |
| Soft delete | Rewrite DELETE as UPDATE of a marker field | [Soft delete](https://insei.github.io/gerpo/features/soft-delete/) |
| Hooks | Before/After for Insert/Update, AfterSelect | [Hooks](https://insei.github.io/gerpo/features/hooks/) |
| Transactions | `gerpo.WithTx(ctx, tx)` / `gerpo.RunInTx` share one tx across every Repository bound to the same context | [Transactions](https://insei.github.io/gerpo/features/transactions/) |
| Cache | Context-scoped cache out of the box, pluggable backend | [Cache](https://insei.github.io/gerpo/features/cache/) |
| Error handling | `WithErrorTransformer` maps gerpo errors to domain errors | [Error transformer](https://insei.github.io/gerpo/features/error-transformer/) |

## Supported adapters

gerpo talks to a database through an `executor.Adapter` — a thin wrapper around an underlying SQL driver. gerpo targets **PostgreSQL** today; all three bundled adapters wrap PostgreSQL drivers:

| Adapter | Package | Wraps driver | Placeholders |
|---|---|---|---|
| pgx v5 | `executor/adapters/pgx5` | `github.com/jackc/pgx/v5` | `$1, $2, …` |
| pgx v4 | `executor/adapters/pgx4` | `github.com/jackc/pgx/v4` | `$1, $2, …` |
| database/sql | `executor/adapters/databasesql` | any `database/sql` driver — pair with a PG driver (`pq`, `pgx/stdlib`) | `?` or `$1` (configurable) |

PG-compatible databases (CockroachDB, MariaDB ≥10.5, SQLite ≥3.35) are likely to work as drop-in — not formally tested. MySQL, MS SQL Server, and older SQLite are **not supported**: gerpo's LIKE `CAST(? AS text)`, `INSERT … RETURNING`, and window-function `COUNT(*) OVER ()` all assume PG. See [`TODO.md`](TODO.md).

Writing a custom adapter is three methods (`ExecContext`, `QueryContext`, `BeginTx`) — see [Adapters](https://insei.github.io/gerpo/features/adapters/) and [adapter internals](https://insei.github.io/gerpo/architecture/adapters-internals/).

## Ideology

1. SQL lives only in the repository configuration.
2. Columns are bound to struct fields through pointers.
3. Entities carry no database markers (no tags, no interfaces).
4. gerpo does not implement relations between entities.
5. gerpo does not modify the database schema.

Details and rationale: [Ideology](https://insei.github.io/gerpo/architecture/ideology/).

## Performance

gerpo uses minimal reflection and pools statement objects to keep allocations under control. Measured against a real pgx v4 pool on the same query:

- `ns/op`: **+8%** compared to the raw driver.
- `B/op` / `allocs/op`: ~2× — the price of generic SQL generation and struct-field mapping.

Against a mock adapter (IO = 0) the relative overhead is larger, but absolute cost per call is ≈0.5–1.5 µs. In a real database, network and query time dwarf the framework overhead. The full mock comparison matrix is produced by `GERPO_BENCH_REPORT=1 go test -run=TestCompareDirectVsGerpo -v ./tests/`.

## Roadmap

**1.0.0**

- [x] Caching engine configuration in the repository builder (#46).
- [ ] New API for configuring virtual columns (current one marked deprecated).

The rest of the API is stable and not expected to change in 1.0.0.

**1.1.0**

- Return inserted IDs and generated timestamps.

## Contributing

- Unit tests: `go test ./...`
- Integration tests (Docker required):

  ```bash
  docker compose -f tests/integration/docker-compose.yml up -d
  GERPO_INTEGRATION_DB_URL="postgres://gerpo:gerpo@localhost:5433/gerpo?sslmode=disable" \
      go test -tags=integration ./tests/integration/...
  ```

- Every PR runs a mock-db benchmark diff via `benchstat` and posts the summary as a PR comment.

More in [Contributing](https://insei.github.io/gerpo/architecture/contributing/).

## License

MIT — see [LICENSE.md](LICENSE.md).
