# Caching internals

The cache is a plug-in. `executor/cache/types.Storage` is the interface; `executor/cache/ctx` is the bundled request-scope implementation.

**Design stance**: gerpo ships a request-scope cache and only a request-scope cache. Distributed / long-lived caching is deliberately out of scope (see the user-facing [Cache](../features/cache.md) page for the rationale). This document describes the internals of the bundled implementation and the contract any `Storage` has to uphold.

## The Storage interface

```go
type Storage interface {
    Get(ctx context.Context, stmt string, args ...any) (any, error)
    Set(ctx context.Context, value any, stmt string, args ...any)
    Clean(ctx context.Context)
}
```

- **Get** returns the cached value or `cache.ErrNotFound`.
- **Set** records a value.
- **Clean** wipes every cached entry reachable through the given context — any write through any repository calls it, so stale reads cannot linger between operations.

The executor is the only caller. `executor/cache.go` wraps `Storage` in three helpers (`get[T]`, `set`, `clean`) that accept a nil storage and no-op — switching the cache off is as simple as not passing `executor.WithCacheStorage`.

## ctx.Cache (bundled implementation)

Source: `executor/cache/ctx/source.go`, `executor/cache/ctx/storage.go`.

The payload lives **inside the request context**. A repo-level `Cache` object has a stable UUID-shaped `key` that lets it partition the context-scoped storage by repo — Get/Set stay isolated between repositories so two repos encoding the same SQL don't cross-contaminate.

```
    ┌────────────────────────────────────────────┐
    │  context.Context                           │
    │                                            │
    │  ctxCacheKey → *cacheStorage {             │
    │      mtx:  sync.Mutex                      │
    │      c:    map[string]map[string]any       │
    │                                            │
    │            ↑               ↑               │
    │   repo A's key    repo B's key             │
    │                                            │
    │  }                                         │
    └────────────────────────────────────────────┘
```

- `ctx.WrapContext(ctx)` installs a `cacheStorage` into the context.
- Repo A's reads/writes go to `storage.c["repo-A-uuid"]`.
- On `Clean`, **every** repo's bucket is cleared — not just the caller's.

`cacheStorage.Get` looks up `modelKey → key → value`. `modelKey` is the repo UUID; `key` is `sql + args`.

## Invalidation

The executor calls `clean(ctx, cacheSource)` after successful `InsertOne`, `Update`, `Delete`. Read operations don't invalidate. Thus:

- A repo-managed mutation always clears the entire per-context cache. Every other repository sharing that context misses on its next read and refills from the database.
- Changes made to the database through a different path (raw SQL, another service) are invisible to the cache until someone writes through gerpo in that context.

### Why wipe everything?

Cross-repo dependencies — virtual columns, JOINs, persistent queries — make it impossible to know statically which cached entries depend on the mutated row. A per-repo clean was the original design and it leaked stale values across repos that shared a JOIN. The request-scope of the cache keeps this over-invalidation cheap: one request typically has a handful of cached reads, and any mutation usually ends the read cycle anyway.

## Cache key

`Cache.Get/Set` build a string key from the SQL statement and its arguments. Until recently this went through `fmt.Sprintf("%s%v", sql, args)`, which allocated a slice of strings plus the formatted copy. Now keys go through a `strings.Builder` with a type switch covering common argument types (`string`, integers, `uuid.UUID`, `[]byte`, `bool`, `nil`) — each of those writes directly into the builder. Uncommon types fall back to `fmt.Fprint`.

Result: 3–5 fewer allocations per cache operation, no change in key identity.

## Thread safety

`cacheStorage` is protected by a `sync.Mutex`. The lock is fine-grained per write and per read — contention is low in practice because most requests touch distinct keys.

## Building your own Storage

Anything satisfying `cache.Storage` works, but any custom implementation has to honour the **wipe-all-on-Clean** contract. A per-key invalidation shim would reintroduce the cross-repo staleness bug the request-scope design exists to prevent.

If you need distributed caching, don't express it through `cache.Storage` — put it at the application layer (HTTP middleware, gRPC interceptor) where business transaction boundaries are explicit. The bundled `ctx.Cache` can run underneath as an intra-request dedup layer.
