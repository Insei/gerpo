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

Full runnable sample lives in [`examples/todo-api/`](examples/todo-api/) — a CRUD REST service with PostgreSQL, goose migrations and docker-compose wiring. Additional end-to-end scenarios are in the [integration tests](tests/integration/).

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

gerpo uses minimal reflection and pools statement objects to keep allocations under control. Two views of the overhead — a mock adapter isolates the framework cost, a real PostgreSQL shows the cost a caller actually experiences with network round-trip in the picture.

**Against real PostgreSQL.** `make bench-report-pg` spins up an isolated `postgres:16` in Docker, applies the bench schema, runs every CRUD op paired (pgx v5 pool vs gerpo repo), and tears the stack down. Sample run on a local machine:

| Op        | Direct ns/op | Gerpo ns/op | × ns | × B  | × allocs |
|-----------|-------------:|------------:|-----:|-----:|---------:|
| GetFirst  | 59 804       | 66 878      | 1.1× | 2.0× | 1.5×     |
| GetList   | 84 030       | 100 375     | 1.2× | 1.2× | 1.1×     |
| Count     | 105 780      | 162 432     | 1.5× | 2.6× | 2.9×     |
| Insert    | 1 607 957    | 1 638 373   | 1.0× | 2.4× | 2.0×     |
| Update    | 1 488 061    | 1 621 205   | 1.1× | 3.1× | 2.6×     |
| Delete    | 58 162       | 63 522      | 1.1× | 2.3× | 2.0×     |

Reads and Delete-on-miss come out at roughly +10 % latency. `INSERT` / `UPDATE` sit at ~1.6 ms per call on a local PG — that is a real fsync on commit, not framework overhead; the gerpo layer contributes ~30 µs on top. `Count` is the outlier at +50 % because a trivial `SELECT count(*) WHERE age >= ?` is so cheap that gerpo's fixed per-call cost is visible as a percentage; it shrinks on non-trivial queries. Allocation ratios reflect the price of generic SQL generation and struct-field mapping.

**Against a mock adapter** (IO = 0, `make bench-report`) the ratios are larger — the framework cost is no longer amortised by network. Per-op absolute cost stays in the 0.5–1.5 µs band, which is what survives on real traffic.

## Static analysis — gerpolint

WHERE operators (`EQ`, `In`, `Contains`, …) take `any`, so the compiler cannot
catch `h.Where().Field(&m.Age).EQ("18")` — field is `int`, argument is a
string — until runtime. gerpo ships a `go/analysis` checker that catches
these mismatches at `go vet` time.

```bash
go install github.com/insei/gerpo/cmd/gerpolint@latest
gerpolint ./...
# …or from a clone:
make lint-gerpolint
```

Rules (`GPL001`..`GPL005`): scalar type mismatch, variadic element mismatch,
string-only operator on non-string field, unresolved field pointer, and
`any`-typed argument. Silence specific lines with `//gerpolint:disable-line`,
`//gerpolint:disable-next-line[=GPL001,…]`, or the
`//gerpolint:disable` / `//gerpolint:enable` block pair.

**Using gerpolint as a golangci-lint plugin.** Drop the repo's
[`.custom-gcl.yml`](.custom-gcl.yml) into your project (pointing
`module: github.com/insei/gerpo`, `import: github.com/insei/gerpo/gerpolintplugin`),
add gerpolint to your linters config, and build a bespoke binary:

```bash
golangci-lint custom         # produces ./bin/custom-gcl with gerpolint embedded
./bin/custom-gcl run ./...
```

```yaml
# .golangci.yml
linters:
  enable: [gerpolint]
  settings:
    custom:
      gerpolint:
        type: module
        settings:
          unresolved-field: skip   # skip | warn | error
          any-arg: warn            # skip | warn | error
          disabled-rules: []       # [GPL001, GPL002, …]
```

## Roadmap

**1.0.0**

- [x] Caching engine configuration in the repository builder (#46).
- [x] New API for configuring virtual columns — `Compute(sql, args...)` replaced `WithSQL`; `Aggregate()` marks aggregate expressions; `Filter(op, spec)` registers per-operator overrides.

The API is now stable and ready for v1.0.0.

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
