# Cache

gerpo can attach a `cache.Storage` to the executor. The built-in implementation is `CtxCache`: a cache scoped to a single `context.Context`. It helps when one business operation fetches the same records multiple times.

## Wiring

```go
import (
    "github.com/insei/gerpo"
    "github.com/insei/gerpo/executor"
    cachectx "github.com/insei/gerpo/executor/cache/ctx"
)

c := cachectx.New() // one instance per repo

repo, _ := gerpo.NewBuilder[User]().
    DB(adapter, executor.WithCacheStorage(c)).
    Table("users").
    Columns(/* … */).
    Build()
```

For each request you must **wrap the ctx**:

```go
reqCtx := cachectx.NewCtxCache(ctx)

// every subsequent call should go through reqCtx
repo.GetFirst(reqCtx, whereByID)
repo.GetFirst(reqCtx, whereByID) // ← hit, served from cache
```

Without `NewCtxCache(ctx)` the cache just does nothing — a warning goes to the log, and the queries themselves still work.

## Behavior

| Operation | Effect |
|---|---|
| `GetFirst`, `GetList`, `Count` | Read and fill the cache by `sql + args` |
| `Insert`, `Update`, `Delete` | **Clear** the repo's cache (`Clean`) |
| External change to the DB | Not observed by the cache — a stale value is served until the next `Insert/Update/Delete` through the repo or until the context ends |

## When it helps

- **N+1 protection:** one business call hits `repo.GetFirst(ctx, id)` from several places — the cache returns the first result.
- **Proxying middleware:** wrap an incoming HTTP request in `NewCtxCache`, and the whole handler/service tree shares the cache.

## When to skip

- You need up-to-date reads more than you need fewer round-trips (e.g. view-after-write).
- A long-running business operation — the cache has no TTL.
- The same gerpo repository is read across different contexts — distinct contexts have independent caches by design.

## Custom cache backend

`executor.WithCacheStorage` accepts any `cache.Storage` — the interface is defined in `executor/cache/types`. You can implement Redis, memcached, or a `sync.Map` with your own policy.

```go
type Storage interface {
    Get(ctx context.Context, stmt string, args ...any) (any, error)
    Set(ctx context.Context, value any, stmt string, args ...any)
    Clean(ctx context.Context)
}
```

## Key performance

Starting with the version that introduced the integration tests, cache keys are built with `strings.Builder` plus a type switch over common parameter types — no `fmt.Sprintf`. That saves 3–5 allocations per cache operation.
