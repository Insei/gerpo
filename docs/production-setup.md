# Production-ready setup

The individual feature pages walk through one concern at a time: [cache](features/cache.md), [tracing](features/tracing.md), [transactions](features/transactions.md), [error transformer](features/error-transformer.md), [adapters](features/adapters.md). This page is the copy-paste starting point that puts them all in one place — pgx v5, goose for schema migrations, OpenTelemetry, request-scope cache, domain error mapping.

It is opinionated on purpose. Substitute pieces to taste; the layering stays the same.

!!! tip "Want a working example?"
    Everything below is assembled into a runnable project under [`examples/todo-api/`](https://github.com/Insei/gerpo/tree/main/examples/todo-api) — a CRUD REST service for a `tasks` table with PostgreSQL, goose migrations and docker-compose wiring. `docker compose up --build` and the API boots on `:8080`. Read it alongside the snippets on this page.

!!! warning "PostgreSQL only"
    All code below assumes PostgreSQL. gerpo's SQL fragments are PG-shaped — see [TODO](https://github.com/insei/gerpo/blob/main/TODO.md) for the multi-dialect backlog.

## The stack

```
┌──────────────────────────────────────────┐
│ HTTP handler / gRPC / CLI command        │  ← your code
├──────────────────────────────────────────┤
│ service layer (domain logic)             │  ← your code
│     • gerpo.RunInTx for atomic work      │
│     • returns domain errors              │
├──────────────────────────────────────────┤
│ gerpo.Repository[T]                      │  ← this library
│     • WithTracer  → OTel spans           │
│     • WithCacheStorage → ctx-scope cache │
│     • WithErrorTransformer → domain errs │
├──────────────────────────────────────────┤
│ executor.Adapter (pgx5)                  │  ← this library
├──────────────────────────────────────────┤
│ pgxpool.Pool                             │  ← github.com/jackc/pgx/v5
├──────────────────────────────────────────┤
│ PostgreSQL 14+                           │
└──────────────────────────────────────────┘
```

Two cross-cutting concerns wrap this stack from outside:

- **Schema migrations** — run once at process start before the pool opens for business.
- **Request-scope cache** — an HTTP middleware wraps every incoming request's `context.Context` so downstream repository calls share the dedup cache.

## 1. Schema migrations (goose)

gerpo does not own the schema — DDL lives in migrations, run separately. A common layout:

```
migrations/
    0001_users.up.sql
    0001_users.down.sql
    0002_orders.up.sql
    …
```

Run goose in `main()` before opening the pool, or (cleaner) as a dedicated `migrate` subcommand on the same binary:

```go
import (
    "context"
    "database/sql"
    "log"

    _ "github.com/jackc/pgx/v5/stdlib"
    "github.com/pressly/goose/v3"
)

func runMigrations(ctx context.Context, dsn string) error {
    db, err := sql.Open("pgx", dsn)
    if err != nil {
        return err
    }
    defer db.Close()

    goose.SetBaseFS(nil) // or embed.FS from your migrations dir
    if err := goose.SetDialect("postgres"); err != nil {
        return err
    }
    return goose.UpContext(ctx, db, "./migrations")
}
```

gerpo connects only after migrations are done. If you prefer `atlas`, `dbmate` or `tern` — swap this step; nothing below cares.

## 2. pgx v5 pool

```go
import (
    "context"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
)

func newPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
    cfg, err := pgxpool.ParseConfig(dsn)
    if err != nil {
        return nil, err
    }
    cfg.MaxConns = 20                      // tune for your load
    cfg.MinConns = 2
    cfg.MaxConnLifetime = 30 * time.Minute
    cfg.MaxConnIdleTime = 5 * time.Minute
    cfg.HealthCheckPeriod = time.Minute

    return pgxpool.NewWithConfig(ctx, cfg)
}
```

Sane starting point for a small service behind a single PG instance. For production numbers, measure — pool size, idle timeouts and lifetime interact with PgBouncer / PG's `max_connections`.

## 3. OpenTelemetry tracer

Shared hook — plug the same tracer into every repository:

```go
import (
    "context"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/trace"

    "github.com/insei/gerpo"
)

func otelTracer() gerpo.Tracer {
    tr := otel.Tracer("gerpo")
    return func(ctx context.Context, span gerpo.SpanInfo) (context.Context, gerpo.SpanEnd) {
        ctx, s := tr.Start(ctx, span.Op,
            trace.WithAttributes(
                attribute.String("db.system", "postgresql"),
                attribute.String("db.sql.table", span.Table),
            ),
        )
        return ctx, func(err error) {
            if err != nil {
                s.RecordError(err)
                s.SetStatus(codes.Error, err.Error())
            }
            s.End()
        }
    }
}
```

See [Tracing](features/tracing.md) for Datadog / custom stacks.

## 4. Error transformer

One transformer per repository — maps `gerpo.ErrNotFound` (and any driver error you care to recognise) to domain errors.

```go
import (
    "errors"

    "github.com/insei/gerpo"
)

var ErrUserNotFound = errors.New("user not found")

func userErrors(err error) error {
    if errors.Is(err, gerpo.ErrNotFound) {
        return ErrUserNotFound
    }
    // PG-specific mapping:
    //   var pgErr *pgconn.PgError
    //   if errors.As(err, &pgErr) && pgErr.Code == "23505" {
    //       return ErrUserAlreadyExists
    //   }
    return err
}
```

The service layer returns `ErrUserNotFound` to HTTP handlers; HTTP handlers map it to 404. gerpo stays invisible above the repo.

## 5. Repository factory

One `Repository[T]` per table. Keep construction in a single function so the options list (tracer, cache, transformer) stays consistent.

```go
import (
    cachectx "github.com/insei/gerpo/executor/cache/ctx"
    "github.com/insei/gerpo"
    "github.com/insei/gerpo/executor"
    "github.com/insei/gerpo/executor/adapters/pgx5"

    "github.com/jackc/pgx/v5/pgxpool"
)

type Repos struct {
    Users  gerpo.Repository[User]
    Orders gerpo.Repository[Order]
    Items  gerpo.Repository[OrderItem]
}

func buildRepos(pool *pgxpool.Pool) (*Repos, error) {
    adapter := pgx5.NewPoolAdapter(pool)
    tracer := otelTracer()

    users, err := gerpo.New[User]().
        Adapter(adapter, executor.WithCacheStorage(cachectx.New())).
        Table("users").
        Columns(func(m *User, c *gerpo.ColumnBuilder[User]) {
            c.Field(&m.ID).ReadOnly().ReturnedOnInsert()
            c.Field(&m.Email)
            c.Field(&m.Name)
            c.Field(&m.CreatedAt).ReadOnly().ReturnedOnInsert()
        }).
        WithTracer(tracer).
        WithErrorTransformer(userErrors).
        Build()
    if err != nil {
        return nil, err
    }

    // orders, items constructed the same way…
    return &Repos{Users: users /* , Orders: …, Items: … */}, nil
}
```

Each repository gets its own `cachectx.New()` storage — the cache is partitioned by repository, and [cross-repo invalidation](features/cache.md#cross-repo-invalidation) is still automatic because the ctx-wrapping middleware binds them to the same request.

## 6. Request-scope cache middleware

The cache only works when the incoming `ctx` has been wrapped with `cachectx.WrapContext`. Do that once per request:

```go
import (
    "net/http"

    cachectx "github.com/insei/gerpo/executor/cache/ctx"
)

func CacheMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := cachectx.WrapContext(r.Context())
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

Every repository call downstream of this middleware now dedupes reads and auto-invalidates on writes, for the lifetime of the request.

!!! info "Distributed cache?"
    `executor.WithCacheStorage` accepts any `cache.Storage`, but the interface has no TTL — gerpo's cache is **request-scope only**. For Redis / memcached, cache at the service layer, above the repository. See [Cache → Distributed caching](features/cache.md#distributed-caching-out-of-scope).

## 7. Transactions

Atomic work goes through `gerpo.RunInTx`. The transaction is propagated via `context.Context`, so every repository invoked with the inner ctx shares it.

```go
import (
    "context"

    "github.com/insei/gerpo"
    "github.com/insei/gerpo/executor/adapters/pgx5"
)

func (s *OrderService) Create(ctx context.Context, order *Order) error {
    return gerpo.RunInTx(ctx, s.adapter, func(ctx context.Context) error {
        if err := s.repos.Orders.Insert(ctx, order); err != nil {
            return err
        }
        _, err := s.repos.Items.InsertMany(ctx, order.Items)
        return err
    })
}
```

- `WithTracer` fires on every repository call inside the tx — the spans are nested in whatever span covers `Create`.
- `WithErrorTransformer` still runs — the service layer sees `ErrUserNotFound`, not `gerpo.ErrNotFound`.
- The cache middleware invalidates automatically after the Insert; the next read inside the same request sees fresh data.

## 8. Putting it together — `main()`

```go
func main() {
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer cancel()

    dsn := os.Getenv("DATABASE_URL")

    // 1. Migrations — before the pool opens.
    if err := runMigrations(ctx, dsn); err != nil {
        log.Fatalf("migrate: %v", err)
    }

    // 2. Pool.
    pool, err := newPool(ctx, dsn)
    if err != nil {
        log.Fatalf("pool: %v", err)
    }
    defer pool.Close()

    // 3. OTel — set up tracer provider here (omitted).
    // 4. Repositories — tracer + cache + error transformer wired inside.
    repos, err := buildRepos(pool)
    if err != nil {
        log.Fatalf("build repos: %v", err)
    }

    svc := NewOrderService(repos, pgx5.NewPoolAdapter(pool))

    mux := http.NewServeMux()
    mux.Handle("/orders", http.HandlerFunc(svc.CreateOrderHandler))

    srv := &http.Server{
        Addr:              ":8080",
        Handler:           CacheMiddleware(mux),
        ReadHeaderTimeout: 5 * time.Second,
    }
    // … graceful shutdown elided …
    log.Fatal(srv.ListenAndServe())
}
```

## Deployment checklist

Before rolling out to production:

- [ ] Migrations run in a separate startup step (or a separate deploy job).
- [ ] `pgxpool` sized against PG's `max_connections` and any PgBouncer in front.
- [ ] Every repository has the same `WithTracer` hook — so traces are uniform.
- [ ] `WithErrorTransformer` set on repositories that need domain error mapping.
- [ ] `CacheMiddleware` wraps every HTTP entry point; background workers call `cachectx.WrapContext` explicitly at the start of each unit of work.
- [ ] All multi-step writes go through `gerpo.RunInTx` — not a bare sequence of `repo.Insert`.
- [ ] OTel exporter, logging and metrics are configured *outside* gerpo; gerpo only emits spans through the `Tracer` hook.
- [ ] [gerpolint](features/static-analysis.md) is wired into CI (standalone or via the golangci-lint plugin) so that `EQ("18")` on an `int` column fails the build rather than the request.

## Related pages

- [Cache](features/cache.md) — scope boundary, cross-repo invalidation.
- [Tracing](features/tracing.md) — Datadog / custom-stack wiring.
- [Transactions](features/transactions.md) — manual `BeginTx`, savepoints.
- [Error transformer](features/error-transformer.md) — what flows through, what doesn't.
- [Adapters](features/adapters.md) — writing custom wrappers (tracing, mocks).
- [Static analysis (gerpolint)](features/static-analysis.md) — catch WHERE-filter type errors at `go vet` time.
