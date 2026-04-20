# TODO — deferred work

This file collects design + implementation tasks that are deliberately out of
scope for the current milestone. Each item has enough context to pick up cold;
none of them block the pre-1.0 cleanup.

---

## Multi-database support (currently: PostgreSQL-only in practice)

gerpo's public messaging suggests "any SQL backend behind a `DBAdapter`", but
the bundled adapters and the SQL fragments we emit are silently
PostgreSQL-shaped. Several places need work before another major dialect can be
treated as a first-class target.

### Known dialect mismatches in emitted SQL

- **LIKE / CONCAT type cast** (`sqlstmt/sqlpart/where.go`): we emit
  `CAST(? AS text)` inside `CONCAT(...)`. The accompanying comment claims this
  is portable to MySQL, but MySQL has no `text` type for `CAST` — the standard
  there is `CAST(? AS CHAR)`. Need to:
  - Add a dialect-aware text-cast helper (e.g. `dialect.TextCast()`).
  - Resolve it through the adapter capability (see below).
  - Add MySQL integration tests covering every LIKE-family operator.

- **`COUNT(*) OVER ()` window function** (`sqlstmt/count.go`): supported in PG
  9.2+, MySQL 8.0+, SQL Server 2005+, SQLite 3.25+. Should work everywhere
  recent enough, but worth verifying on a real MySQL box.

- **Boolean literals**: PG accepts `TRUE`/`FALSE`; MySQL prefers `1`/`0` for
  performance and SQL Server requires `1`/`0`. Currently we rely on driver-side
  parameter binding so this is implicit, but custom virtual-column filters may
  trip on it.

### RETURNING

- **MySQL has no `RETURNING`** (MariaDB 10.5+ does). Fallback options
  considered and *deferred*:
  - `LastInsertId()` from `sql.Result` — covers only single auto-increment PK
    inserts; doesn't help with UUID DEFAULT, `created_at` triggers, or UPDATE.
  - Read-back via a separate `SELECT` keyed on PK — needs gerpo to know which
    column is the PK, doesn't help when PK is also DB-generated.
  - User-supplied read-back via `WithAfterInsert` hook (now that hooks return
    error) — this is the recommended workaround today.

  When MySQL support lands, the cleanest path is probably the
  hook-based read-back wired into a stable adapter capability so `Build()` does
  not refuse to build a repo that uses `ReturnedOnInsert/Update`.

- **SQL Server uses `OUTPUT` clause**, not `RETURNING`. Need a dialect helper
  that emits the right syntax (`OUTPUT INSERTED.col` for INSERT, slightly
  different for UPDATE).

### Placeholder formats

- Currently handled — pgx adapters use `placeholder.Dollar`, database/sql is
  configurable. SQL Server uses `@p1`/`@p2` style — would need a new
  `placeholder.AtName` format.

### Dialect detection / adapter capability

- Today the only adapter capability is `ReturningCapable` (added with the
  RETURNING feature). Cleaner long-term shape: a single `Dialect()` method on
  `Adapter` returning a `Dialect` value with named capabilities + SQL helpers.
  Defer until we actually have a non-PG dialect to support.

### Bundled adapters

- We have pgx5, pgx4, database/sql. When MySQL/MariaDB/SQL Server support
  lands, the adapter README/index should mark which dialects each adapter
  actually targets in production.

### Integration tests

- `tests/integration/` only has Postgres infra (docker-compose, schema).
  Needs:
  - Parallel Docker compose stack for MySQL / MariaDB / SQL Server.
  - `forEachAdapter` loop extended to spin per-dialect repositories.
  - Dialect-specific test paths skipped when the adapter under test does not
    support the feature (e.g. a `t.Skip(...)` matrix per dialect capability).

### Documentation

- `docs/features/adapters.md` and `docs/architecture/adapters-internals.md`
  need a "Supported dialects" matrix once we land more than PG.
- README's "Supported drivers" section should be split into "drivers" vs
  "dialects" — the underlying confusion is exactly the one we're trying to
  resolve in code.
