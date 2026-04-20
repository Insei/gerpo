# Caching internals

The cache is a plug-in. `executor/cache/types.Storage` is the interface; `executor/cache/ctx` is the bundled context-scoped implementation.

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
- **Clean** wipes the cache (the executor calls it after INSERT/UPDATE/DELETE so stale reads don't linger).

The executor is the only caller. `executor/cache.go` wraps `Storage` in three helpers (`get[T]`, `set`, `clean`) that accept a nil storage and no-op — so switching the cache off is as simple as not passing `executor.WithCacheStorage`.

## Cache

Source: `executor/cache/ctx/source.go`, `executor/cache/ctx/storage.go`.

The idea: the cache payload lives **inside the request context**. A repo-level `Cache` object has a stable UUID-shaped `key` that lets it partition the context-scoped storage by repo.

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
- On `Clean` only the repo's bucket is cleared, not the whole tree.

`cacheStorage.Get` looks up `modelKey → key → value`. `modelKey` is the repo UUID; `key` is `sql + args`.

## Cache key

`Cache.Get/Set` build a string key from the SQL statement and its arguments. Until recently this went through `fmt.Sprintf("%s%v", sql, args)`, which allocated a slice of strings plus the formatted copy. Now keys go through a `strings.Builder` with a type switch covering common argument types (`string`, integers, `uuid.UUID`, `[]byte`, `bool`, `nil`) — each of those writes directly into the builder. Uncommon types fall back to `fmt.Fprint`.

Result: 3–5 fewer allocations per cache operation, no change in key identity.

## Invalidation

The executor calls `clean(ctx, cacheSource)` after successful `InsertOne`, `Update`, `Delete`. Read operations don't invalidate. Thus:

- a repo-managed mutation always clears the repo's cache bucket for the current request context;
- changes made to the database through a different path (raw SQL, another repo, an external service) are invisible to the cache until someone writes through the repo in that context.

## Thread safety

`cacheStorage` is protected by a `sync.Mutex`. The lock is fine-grained per write and per read — contention is low in practice because most requests touch distinct keys.

## Building your own Storage

Anything satisfying `cache.Storage` works. Common candidates:

- Redis — for cross-instance sharing within a trace/tenant.
- `sync.Map` with TTL — for a process-wide cache, but mind invalidation semantics.
- `cache.NewModelBundle` — already bundled; combines several storages into one (e.g. ctx + redis).
