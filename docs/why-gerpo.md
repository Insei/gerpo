# Why gerpo?

Go has a healthy data-access ecosystem — full-blown ORMs, code generators, query builders, thin wrappers. gerpo occupies a specific niche: a **type-safe repository pattern with pluggable SQL adapters and no schema management**. This page lays out where it fits, what it gives up, and how it compares to the closest alternatives.

## The 30-second pitch

- One **declarative configuration** per entity wires struct fields to columns through pointers — `c.Field(&m.Email)` — so renames are a refactor, not a search-and-replace through string tags.
- Six methods per repository (`GetFirst`, `GetList`, `Count`, `Insert`, `Update`, `Delete`) plus `InsertMany` cover the everyday CRUD; everything else (joins, soft-delete, virtual columns, hooks, caching, tracing) is opt-in.
- **PostgreSQL-only** today. Three bundled adapters (`pgx5`, `pgx4`, `database/sql`) all wrap PG drivers. Other dialects (MySQL, MS SQL Server, older SQLite) are on the backlog, not on the main path.
- **Static type-checker included.** [`gerpolint`](features/static-analysis.md) catches `Field(&m.Age).EQ("18")` and friends at `go vet` time — shipped as a standalone binary **and** as a `golangci-lint` v2 module plugin.
- **Not** an ORM. No migrations, no relations, no struct tags. Schema management is your problem (`golang-migrate`, `goose`, `atlas`, …).
- **API stable** as of v1.0.0 (2026-04-20). Breaking changes go through SemVer majors.

## When to pick gerpo

Pick gerpo when you want:

- A clear, type-safe boundary between business code and SQL, backed by a `go/analysis` checker that enforces the rule at build time.
- Predictable allocations and SQL generation — `make bench-report` shows the overhead per operation.
- You run on PostgreSQL (or a PG-compatible database — CockroachDB, MariaDB ≥10.5, SQLite ≥3.35) and want multiple PG drivers behind one interface (pgx v5, pgx v4, `database/sql` + `pq` or `pgx/stdlib`).
- Per-request caching that just turns on (`Cache`).
- An OpenTelemetry-style tracing hook without forcing OTel as a dependency.
- A small, readable codebase you can fork or wrap.

## When **not** to pick gerpo

Skip it if:

- You want migrations bundled with your data layer — pick **GORM** or **ent** instead.
- You want navigation properties / lazy loading (`user.Posts`, `post.Comments`) — `gerpo` deliberately doesn't provide them.
- You need to talk to **MySQL, MS SQL Server, or older SQLite**. gerpo emits PostgreSQL-shaped SQL and does not abstract the dialect today — see [TODO](https://github.com/insei/gerpo/blob/main/TODO.md) for the backlog.
- Your team already runs on raw SQL and wants compile-time-checked queries from `.sql` files — pick **sqlc**.
- You only need a thin marshalling layer over `database/sql` — pick **sqlx**.

## Feature matrix

|  | gerpo | [GORM](https://gorm.io/) | [ent](https://entgo.io/) | [bun](https://bun.uptrace.dev/) | [sqlc](https://sqlc.dev/) | [sqlx](https://github.com/jmoiron/sqlx) |
|---|---|---|---|---|---|---|
| **Approach** | Repository + SQL config | Active Record / ORM | Schema DSL + codegen | ORM-lite | SQL-first codegen | `database/sql` wrapper |
| **Schema source** | Go config (pointers) | Struct tags | Go DSL | Struct tags | `.sql` files | None |
| **Type-safe queries** | ✓ (generics) | partial | ✓ | partial | ✓ | ✗ |
| **Code generation** | ✗ | ✗ | ✓ | ✗ | ✓ | ✗ |
| **Migrations** | ✗ (external) | ✓ | ✓ | ✓ | ✗ | ✗ |
| **Relations / navigation** | ✗ | ✓ | ✓ | ✓ | ✗ | ✗ |
| **Struct tags required** | ✗ | ✓ | ✗ | ✓ | ✗ | ✓ |
| **Supported databases** | PostgreSQL only | multi-dialect | multi-dialect | multi-dialect | multi-dialect | any (via `database/sql`) |
| **Pluggable drivers (within dialect)** | ✓ (pgx5, pgx4, database/sql) | ✓ | ✓ | ✓ | ✓ | ✓ |
| **Soft delete built-in** | ✓ | ✓ | ✓ | ✓ | ✗ | ✗ |
| **Per-request cache** | ✓ (`Cache`) | plugin | ✗ | ✗ | ✗ | ✗ |
| **Tracing hook** | ✓ | plugin | hook | hook | ✗ | ✗ |
| **Hooks (Before/After)** | ✓ | ✓ | ✓ | ✓ | ✗ | ✗ |
| **Custom SQL escape hatch** | ✓ (callbacks) | ✓ (`Raw()`) | ✓ | ✓ | n/a (everything is SQL) | ✓ |
| **Reflection** | only at config (fmap + unsafe offsets) | runtime, every call | none (generated code) | runtime | none | minimal |
| **Static checker (go vet-time)** | ✓ ([`gerpolint`](features/static-analysis.md), also a golangci-lint plugin) | ✗ | ✓ (generated code is typed) | ✗ | ✓ (generated signatures) | ✗ |

## Strengths

- **Pointer-based mapping is refactor-proof for deletes and type changes.** Remove a field and every `Field(&m.X)` in the codebase stops compiling; change its type and the operator-level check tightens. (Renames still default to `snake_case` of the Go name — pin with `.WithName("age")` on stable columns.)
- **Type-safety extends past `go build`.** [`gerpolint`](features/static-analysis.md) — a `go/analysis` checker shipped with gerpo — catches `Field(&m.Age).EQ("18")` at `go vet` time (field is `int`, argument is `string`), flags `Contains` on non-string fields, and reasons about `In([]any{...})` element types. Available as a standalone binary (`gerpolint ./...`) and as a **golangci-lint v2 module plugin** with a ready-to-use [`.custom-gcl.yml`](https://github.com/insei/gerpo/blob/main/.custom-gcl.yml).
- **No surprise SQL.** Every JOIN, GROUP BY, virtual column and persistent filter is in one `WithQuery(...)` block per repository. There is no hidden auto-load that spawns a second query behind your back.
- **Three adapters, one base.** Driver-specific code is a `Driver{Exec,Query,BeginTx}` + `TxDriver{…}` pair (a few dozen lines). The placeholder rewrite, the transaction state machine and `RollbackUnlessCommitted` semantics live once in `executor/adapters/internal`.
- **Cache and tracing are first-class but opt-in.** `WithCacheStorage` and `WithTracer` take small interfaces — implement them with whatever your stack already uses (Redis, OTel, Datadog, …) without dragging dependencies into gerpo.
- **Battle-tested in CI.** Every PR runs lint (including `gerpolint` via `golangci-lint custom`), race-detector unit tests, integration tests against a real PostgreSQL service container on three drivers, and a [benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat) overhead diff. See [Contributing](architecture/contributing.md).

## Weaknesses

- **No migrations.** Pick a separate tool — `golang-migrate`, `goose`, `atlas`. The `examples/todo-api/` walkthrough wires goose with `//go:embed`.
- **No relations.** Many-to-one / one-to-many fan-outs are explicit calls — write a `FindPostsByUser(ctx, userID)` rather than `user.Posts`.
- **PostgreSQL-only.** Every adapter wraps a PG driver; the emitted SQL assumes PG (`$1` placeholders, `RETURNING`, window-function pagination, `CAST(? AS text)` in LIKE). Multi-dialect is on the backlog, not the main path.
- **Renames shift column names by default.** `snake_case(Go field name)` is the convention — pin a stable name with `.WithName(...)` on every production column if this matters.
- **Generic boilerplate.** Every repository carries a `[TModel any]` parameter. With many entities you end up with many small `gerpo.Repository[Foo]` typed values. This is the price of compile-time safety.
- **Reflection footprint.** `fmap` walks struct layout once at builder time using `unsafe.Offsetof`; the per-request hot path is pointer arithmetic, not reflection. But: pointer-based field resolution adds ~12 allocations per `GetFirst` over a raw `pgx` call (≈+1 µs). In a real query that figure is noise; in a tight benchmark loop it shows up.

## Performance

`make bench-report` runs every CRUD operation twice — once against a mock backend directly, once through gerpo — and prints a comparison table. Headline numbers (post-optimisation, mock backend):

| Op | Direct ns/op | Gerpo ns/op | Direct allocs | Gerpo allocs |
|---|---:|---:|---:|---:|
| GetFirst | ~200 | ~1300 | 5 | 17 |
| GetList (10 rows) | ~1100 | ~3000 | 21 | 34 |
| Count | ~100 | ~750 | 4 | 10 |
| Insert | ~140 | ~700 | 5 | 14 |
| Update | ~75 | ~1400 | 3 | 23 |
| Delete | ~70 | ~870 | 3 | 17 |

Those ratios look scary in isolation. **In a real database round-trip** (50 µs locally, 500 µs over the network) gerpo's ~1 µs overhead is 0.2–2% — the README's "+8% ns/op" measurement against a real `pgx v4` pool matches that ballpark. The allocation numbers matter more for GC pressure under high RPS than for raw latency.

## Closest alternatives — when each fits better

- **GORM.** You want everything in one box: schema, migrations, relations, hooks. You're fine with the runtime overhead and the occasional surprise from active-record semantics.
- **ent.** Your domain has a deeply connected graph (users → orgs → projects → issues → comments) and you want the compiler to enforce the traversal. You don't mind running a code generator.
- **bun.** You want most of GORM's ergonomics with a smaller surface and explicit relationships.
- **sqlc.** Your team writes SQL by hand and wants compile-time-checked, hand-tuned queries with generated Go signatures. You don't want a query DSL.
- **sqlx.** You want `database/sql` plus row-to-struct marshalling and nothing else.
- **gerpo.** You want the repository pattern as a first-class concept — pointer-based wiring, type-safe per-operation helpers, three adapters, no schema management — and you handle migrations elsewhere.

## Reading list

- [Get started](index.md) — install + 30-line example.
- [Features](features/index.md) — every capability with code samples.
- [Static analysis (gerpolint)](features/static-analysis.md) — what the checker catches, which rules exist, how to wire the golangci-lint plugin.
- [Production-ready setup](production-setup.md) — pgx v5 + goose + OpenTelemetry + cache + domain errors.
- [Runnable example](https://github.com/Insei/gerpo/tree/main/examples/todo-api) — CRUD REST service (~350 LoC) with goose migrations and docker-compose.
- [Architecture](architecture/index.md) — internals for contributors.
- [API reference](https://pkg.go.dev/github.com/insei/gerpo) — runnable godoc examples next to every method.
