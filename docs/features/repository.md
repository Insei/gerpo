# Repository builder

A repository is assembled through the fluent `gerpo.New[T]()`. The chain `DB → Table → Columns → Build` is mandatory. Everything else is optional `With*` steps between `Columns` and `Build`.

## Minimal repository

```go
repo, err := gerpo.New[User]().
    DB(adapter).
    Table("users").
    Columns(func(m *User, c *gerpo.ColumnBuilder[User]) {
        c.Field(&m.ID).OmitOnUpdate()
        c.Field(&m.Name)
    }).
    Build()
```

`Build()` returns `gerpo.Repository[User]` — a thread-safe object, a single instance serves the whole application.

## Full chain of options

```go
repo, err := gerpo.New[User]().
    DB(adapter, executor.WithCacheStorage(ctxCache)).      // (1) adapter + executor options
    Table("users").                                         // (2) table name
    Columns(colsFn).                                        // (3) column description
    WithQuery(persistentFn).                                // (4) persistent conditions
    WithSoftDeletion(softFn).                               // (5) DELETE replacement
    WithBeforeInsert(beforeFn).                             // (6) hooks
    WithAfterInsert(afterFn).
    WithBeforeUpdate(beforeUpdFn).
    WithAfterUpdate(afterUpdFn).
    WithAfterSelect(afterSelectFn).
    WithErrorTransformer(mapErr).                           // (7) error mapping
    Build()
```

1. **DB** — attach a driver adapter. List of bundled adapters: [Adapters](adapters.md). The second argument accepts `executor.Option` (e.g., `WithCacheStorage`).
2. **Table** — physical table name.
3. **Columns** — mandatory column description. [Columns](columns.md).
4. **WithQuery** — persistent filters/joins/groupings applied to every request. [Persistent queries](persistent-queries.md).
5. **WithSoftDeletion** — rewrite DELETE as UPDATE. [Soft delete](soft-delete.md).
6. **With{Before,After}{Insert,Update}**, **WithAfterSelect** — hooks. Multiple calls stack. [Hooks](hooks.md).
7. **WithErrorTransformer** — run every error through a mapping function. [Error transformer](error-transformer.md).

## Methods of the finished repository

```go
type Repository[TModel any] interface {
    GetFirst(ctx context.Context, qFns ...func(m *TModel, h query.GetFirstHelper[TModel])) (*TModel, error)
    GetList(ctx context.Context, qFns ...func(m *TModel, h query.GetListHelper[TModel])) ([]*TModel, error)
    Count(ctx context.Context, qFns ...func(m *TModel, h query.CountHelper[TModel])) (uint64, error)
    Insert(ctx context.Context, model *TModel, qFns ...func(m *TModel, h query.InsertHelper[TModel])) error
    Update(ctx context.Context, model *TModel, qFns ...func(m *TModel, h query.UpdateHelper[TModel])) (int64, error)
    Delete(ctx context.Context, qFns ...func(m *TModel, h query.DeleteHelper[TModel])) (int64, error)
    Tx(tx executor.Tx) Repository[TModel]
    GetColumns() types.ColumnsStorage
}
```

Every method is described in [CRUD operations](crud.md).

## Thread safety

A repository is safe to share across goroutines. Internally, statement objects are reused through `sync.Pool`, but each call takes its own instance from the pool, so there are no races.

## Build errors

`Build()` returns an error if:

- `DB` or `Table` is not set;
- `Columns` contains no columns;
- a virtual column lacks `WithSQL`;
- `WithSoftDeletion` references a field that is not declared as a column.
