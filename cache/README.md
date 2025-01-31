# Cache
The Cache package allows you to retrieve a cached value by using an SQL query and its arguments.

## Architecture
Each cache store must implement the Source interface:
```go
type Source interface {
    Clean(ctx context.Context)
    Get(ctx context.Context, statement string, statementArgs ...any) (any, error)
    Set(ctx context.Context, cache any, statement string, statementArgs ...any) error
}
```
- Each cache source and cache bundle must be unique to a specific model.
- The Clean method is called whenever an update or insert SQL query is executed on a model.
- Cache Get/Set operations occur only when executing SELECT queries.

## Supported caches
- Context-based cache (stores cache data in context; for example, it can be used with HTTP middleware). Refer to the "ctx" package.
- CacheBundle – combines multiple cache sources into a single bundle.

### TODO
- Add asynchronous cache retrieval from sources for CacheBundle.