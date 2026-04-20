# Why gerpo?

Go has a healthy data-access ecosystem — full-blown ORMs, code generators, query builders, thin wrappers. gerpo occupies a specific niche: a **type-safe repository pattern with pluggable SQL adapters and no schema management**. This page lays out where it fits, what it gives up, and how it compares to the closest alternatives.

## The 30-second pitch

- One **declarative configuration** per entity wires struct fields to columns through pointers — `c.Field(&m.Email)` — so renames are a refactor, not a search-and-replace through string tags.
- Six methods per repository (`GetFirst`, `GetList`, `Count`, `Insert`, `Update`, `Delete`) cover the everyday CRUD; everything else (joins, soft-delete, virtual columns, hooks, caching, tracing) is opt-in.
- Three driver adapters (`pgx5`, `pgx4`, `database/sql`) all sit behind a 3-method `Adapter` interface — bring your own driver in ~50 lines.
- **Not** an ORM. No migrations, no relations, no struct tags. Schema management is your problem (`golang-migrate`, `goose`, `atlas`, …).

## When to pick gerpo

Pick gerpo when you want:

- A clear, type-safe boundary between business code and SQL.
- Predictable allocations and SQL generation — `make bench-report` shows the overhead per operation.
- Multiple drivers behind one interface (microservices on PostgreSQL today, easy to add ClickHouse / SQLite tomorrow).
- Per-request caching that just turns on (`Cache`).
- An OpenTelemetry-style tracing hook without forcing OTel as a dependency.
- A small, readable codebase you can fork or wrap.

## When **not** to pick gerpo

Skip it if:

- You want migrations bundled with your data layer — pick **GORM** or **ent** instead.
- You want navigation properties / lazy loading (`user.Posts`, `post.Comments`) — `gerpo` deliberately doesn't provide them.
- Your team already runs on raw SQL and wants compile-time-checked queries from `.sql` files — pick **sqlc**.
- You only need a thin marshalling layer over `database/sql` — pick **sqlx**.
- You can't tolerate an API that is still pre-1.0 — gerpo is on the road to 1.0 with a stated stable subset, but the deprecated virtual-column API is still in flux.

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
| **Pluggable drivers** | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| **Soft delete built-in** | ✓ | ✓ | ✓ | ✓ | ✗ | ✗ |
| **Per-request cache** | ✓ (`Cache`) | plugin | ✗ | ✗ | ✗ | ✗ |
| **Tracing hook** | ✓ | plugin | hook | hook | ✗ | ✗ |
| **Hooks (Before/After)** | ✓ | ✓ | ✓ | ✓ | ✗ | ✗ |
| **Custom SQL escape hatch** | ✓ (callbacks) | ✓ (`Raw()`) | ✓ | ✓ | n/a (everything is SQL) | ✓ |
| **Lines of code** | ~3k | ~50k | ~80k | ~30k | n/a | ~3k |
| **Reflection** | only at config (fmap + unsafe offsets) | runtime, every call | none (generated code) | runtime | none | minimal |

The "Lines of code" row is rough but conveys the shape: gerpo is closer to sqlx in size and to ent in API ergonomics.

## Strengths

- **Pointer-based mapping is refactor-proof.** Rename a field — the compiler tells you everywhere it's wired into a column. Tag-based schemas only break at runtime.
- **No surprise SQL.** Every JOIN, GROUP BY, virtual column and persistent filter is in one `WithQuery(...)` block per repository. There is no hidden auto-load that spawns a second query behind your back.
- **Three adapters, one base.** Driver-specific code is a `Driver{Exec,Query,BeginTx}` + `TxDriver{…}` pair (a few dozen lines). The placeholder rewrite, the transaction state machine and `RollbackUnlessCommitted` semantics live once in `executor/adapters/internal`.
- **Cache and tracing are first-class but opt-in.** `WithCacheStorage` and `WithTracer` take small interfaces — implement them with whatever your stack already uses (Redis, OTel, Datadog, …) without dragging dependencies into gerpo.
- **Pre-1.0 already battle-tested in CI.** Every PR runs lint, race-detector unit tests, integration tests against a real PostgreSQL service container on three drivers, and a [benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat) overhead diff. See [Contributing](architecture/contributing.md).

## Weaknesses

- **No migrations.** You will need a separate tool. We recommend `golang-migrate`, `goose`, or `atlas`.
- **No relations.** Many-to-one / one-to-many fan-outs are explicit calls — write a `FindPostsByUser(ctx, userID)` rather than `user.Posts`.
- **Pre-1.0 API surface.** The virtual-column configuration API is marked deprecated and will be replaced in 1.0.0. The rest of the API is stable per the README roadmap.
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
- [Architecture](architecture/index.md) — internals for contributors.
- [API reference](https://pkg.go.dev/github.com/insei/gerpo) — runnable godoc examples next to every method.
