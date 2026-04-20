# Cache

gerpo ships a **request-scope** deduplication cache. It helps when one business operation fetches the same records multiple times — same SQL + same args → same result served from memory, no driver round-trip.

**Scope boundary, intentional**: the cache lives inside `context.Context` and dies with it. There is no TTL, no cross-request sharing, and no distributed backend bundled with gerpo. For application-wide or distributed caching — put that layer above gerpo (HTTP middleware, gRPC interceptor, a service-layer wrapper).

## Wiring

```go
import (
    "github.com/insei/gerpo"
    "github.com/insei/gerpo/executor"
    cachectx "github.com/insei/gerpo/executor/cache/ctx"
)

c := cachectx.New() // one instance per repo

repo, _ := gerpo.New[User]().
    Adapter(adapter, executor.WithCacheStorage(c)).
    Table("users").
    Columns(/* … */).
    Build()
```

For each request you must **wrap the ctx**:

```go
reqCtx := cachectx.WrapContext(ctx)

// every subsequent call should go through reqCtx
repo.GetFirst(reqCtx, whereByID)
repo.GetFirst(reqCtx, whereByID) // ← hit, served from cache
```

Without `WrapContext(ctx)` the cache just does nothing — a warning goes to the log, and the queries themselves still work.

## Behavior

| Operation | Effect |
|---|---|
| `GetFirst`, `GetList`, `Count` | Read and fill the cache by `sql + args` |
| `Insert`, `Update`, `Delete` | **Wipe the entire per-context cache** (every repository sharing the context) |
| External change to the DB | Not observed by the cache — a stale value is served until some repository writes through the context or until the context ends |

### Cross-repo invalidation

A write through any repository wipes **every** cache entry in the current context, not just the writing repo's bucket. That is the only safe default when repositories can share results through virtual columns or JOINs — gerpo cannot statically know whether `postsRepo`'s cached result depends on a users row that `usersRepo` just mutated.

If a heavier-grained invalidation is fine for your workload — keep the cache hot across repositories, invalidate manually when you have to — lift the cache to the application layer and drive it yourself.

## When it helps

- **N+1 protection:** one handler hits `repo.GetFirst(ctx, id)` from several places — the cache returns the first result.
- **Proxying middleware:** wrap an incoming HTTP request in `WrapContext`, and the whole handler/service tree shares the cache.

## When to skip

- View-after-write patterns where every read must see the freshest state.
- Long-running tasks that span many logical operations — request scope is the wrong granularity.
- Reads across different contexts (cron + HTTP, for instance) — distinct contexts have independent caches by design.

## Distributed caching — out of scope

`executor.WithCacheStorage` accepts any `cache.Storage`, so a Redis-backed implementation would compile. It is **not a path gerpo supports** — the `Storage` interface has no TTL, no eviction policy, and invalidation assumes a single process. Fitting Redis into that shape invites silent correctness bugs (OOM without TTL, stale reads across pods).

Recommended pattern instead: cache at the **application layer** where you know the business transaction boundaries, and let gerpo talk to the driver directly. The request-scope cache still complements that layer — it deduplicates reads within one handler.

## Key performance

Cache keys are built with `strings.Builder` plus a type switch over common parameter types (`string`, integers, `uuid.UUID`, `[]byte`, `bool`, `nil`) — no `fmt.Sprintf`. That saves 3–5 allocations per cache operation.
