# CRUD operations

The `Repository[T]` interface provides six methods: `GetFirst`, `GetList`, `Count`, `Insert`, `Update`, `Delete`. Each one accepts a `context.Context` and a variadic list of query functions that configure a single call.

!!! tip "Reusable per-operation helpers"
    Every per-operation helper (`GetFirstHelper`, `GetListHelper`, `CountHelper`, `InsertHelper`, `UpdateHelper`, `DeleteHelper`) is composed from small contracts in the `query` package: `Filterable` (`Where()`), `Sortable` (`OrderBy()`), `Excludable` (`Exclude/Only`), `Pageable` (`Page/Size`). You can write reusable middleware-style helpers against these narrow interfaces:

    ```go
    func applyTenant(h query.Filterable, tenantID uuid.UUID) {
        h.Where().Field(...).EQ(tenantID)
    }

    repo.GetFirst(ctx, func(m *User, h query.GetFirstHelper[User]) { applyTenant(h, tid) })
    repo.Update(ctx,  &u, func(m *User, h query.UpdateHelper[User])  { applyTenant(h, tid) })
    ```

## GetFirst

Return the first matching record.

```go
u, err := repo.GetFirst(ctx, func(m *User, h query.GetFirstHelper[User]) {
    h.Where().Field(&m.Email).EQ("alice@example.com")
    h.OrderBy().Field(&m.CreatedAt).DESC()
})
```

If no row is found, the error is `gerpo.ErrNotFound`:

```go
if errors.Is(err, gerpo.ErrNotFound) { /* … */ }
```

## GetList

Return a slice of records. An empty result is an empty slice and no error.

```go
users, err := repo.GetList(ctx, func(m *User, h query.GetListHelper[User]) {
    h.Where().Field(&m.Age).GTE(18)
    h.OrderBy().Field(&m.Name).ASC()
    h.Page(2).Size(50)
})
```

See [Ordering & pagination](order-pagination.md) for details on `Page`/`Size`.

## Count

Returns a `uint64`.

```go
n, err := repo.Count(ctx, func(m *User, h query.CountHelper[User]) {
    h.Where().Field(&m.Age).GTE(18)
})
```

## Insert

Inserts a single record. Mutates the model through `WithBeforeInsert` / `WithAfterInsert` hooks (see [Hooks](hooks.md)).

```go
u := &User{ID: uuid.New(), Name: "Bob", Age: 25, CreatedAt: time.Now()}
err := repo.Insert(ctx, u)
```

`Exclude`/`Only` narrows the column set (handy when the database should apply a `DEFAULT`):

```go
repo.Insert(ctx, u, func(m *User, h query.InsertHelper[User]) {
    h.Exclude(&m.CreatedAt) // let the database default NOW()
})
```

`Returning(fields...)` overrides the repo-level `RETURNING` set for this single
call (default: columns marked with `ReturnedOnInsert()` — see
[Columns → Server-generated values](columns.md#server-generated-values-returning)):

```go
// only the ID comes back, not CreatedAt
repo.Insert(ctx, u, func(m *User, h query.InsertHelper[User]) {
    h.Returning(&m.ID)
})

// disable RETURNING entirely for this call
repo.Insert(ctx, u, func(m *User, h query.InsertHelper[User]) {
    h.Returning()
})
```

## InsertMany

Bulk-inserts a slice as a single multi-row `INSERT ... VALUES (...), (...), ...`. The call is transparently chunked at PostgreSQL's 65535-placeholder limit, so arbitrarily large slices are safe. Returns the total number of rows written.

```go
posts := []*Post{
    {UserID: u.ID, Title: "one"},
    {UserID: u.ID, Title: "two"},
    {UserID: u.ID, Title: "three"},
}
n, err := repo.InsertMany(ctx, posts)
```

An empty slice is a no-op: `(0, nil)` with no SQL, no hooks.

`Exclude`/`Only` and `Returning` behave the same as on the single-row `Insert` and apply to **every** row in the batch uniformly — there is no per-row override:

```go
repo.InsertMany(ctx, posts, func(m *Post, h query.InsertManyHelper[Post]) {
    h.Exclude(&m.CreatedAt)  // let the DB default NOW() for every row
    h.Returning(&m.ID)       // pull only IDs back
})
```

When `RETURNING` is active, scanned values are written back into each element of the slice by position.

!!! warning "Atomicity across chunks is the caller's job"
    If a slice exceeds the placeholder budget, `InsertMany` splits it into several SQL statements. A failure mid-batch leaves rows written by prior chunks in place. Wrap the call in `gerpo.RunInTx` if you need all-or-nothing.

Batch-specific hooks (`WithBeforeInsertMany` / `WithAfterInsertMany`) see the full slice in one call — useful for cascading children in one batched child `InsertMany` instead of N serial ones. See [Hooks](hooks.md).

## Update

Updates records by WHERE. Returns the number of affected rows. When zero rows match, returns `gerpo.ErrNotFound`.

```go
u.Name = "Bob The Builder"
affected, err := repo.Update(ctx, u, func(m *User, h query.UpdateHelper[User]) {
    h.Where().Field(&m.ID).EQ(u.ID)
})
```

`Only`/`Exclude` let you update a subset of fields ([Exclude & Only](exclude-only.md)):

```go
repo.Update(ctx, u, func(m *User, h query.UpdateHelper[User]) {
    h.Where().Field(&m.ID).EQ(u.ID)
    h.Only(&m.Name) // SET name = ?, and nothing else
})
```

`Returning(fields...)` works on Update too — same semantics as on Insert
(default: columns marked `ReturnedOnUpdate()`; explicit list narrows; empty
call disables RETURNING for this call):

```go
repo.Update(ctx, u, func(m *User, h query.UpdateHelper[User]) {
    h.Where().Field(&m.ID).EQ(u.ID)
    h.Returning(&m.UpdatedAt) // bring back only the trigger-set timestamp
})
```

## Delete

Deletes records by WHERE. If the repo was configured with `WithSoftDeletion`, this is rewritten as an UPDATE instead ([Soft delete](soft-delete.md)). When zero rows match, returns `gerpo.ErrNotFound`.

```go
n, err := repo.Delete(ctx, func(m *User, h query.DeleteHelper[User]) {
    h.Where().Field(&m.ID).EQ(u.ID)
})
```

!!! warning "Delete without WHERE wipes the table"
    The repo does not block an unconditional `Delete` unless a persistent query puts a WHERE in front. Always pass a WHERE explicitly.

## Error semantics

| Method | ErrNotFound when |
|---|---|
| `GetFirst` | no rows returned |
| `Update` | `RowsAffected == 0` |
| `Delete` | `RowsAffected == 0` (including the UPDATE from soft delete) |
| `GetList`, `Count`, `Insert` | **never** |

Any other error (FK, unique, syntax, network) is returned as-is and passed through [`WithErrorTransformer`](error-transformer.md) if configured.
