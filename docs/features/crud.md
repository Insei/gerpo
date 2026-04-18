# CRUD operations

The `Repository[T]` interface provides six methods: `GetFirst`, `GetList`, `Count`, `Insert`, `Update`, `Delete`. Each one accepts a `context.Context` and a variadic list of query functions that configure a single call.

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
