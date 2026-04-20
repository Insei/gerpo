# Soft delete

`WithSoftDeletion(fn)` turns a physical DELETE into an UPDATE of selected fields. The "mark, don't drop" pattern — useful to preserve data and keep foreign keys intact.

## Setup

```go
.Columns(func(m *User, c *gerpo.ColumnBuilder[User]) {
    c.Field(&m.ID)
    c.Field(&m.DeletedAt).OmitOnInsert()
}).
WithSoftDeletion(func(m *User, b *gerpo.SoftDeletionBuilder[User]) {
    b.Field(&m.DeletedAt).SetValueFn(func(ctx context.Context) any {
        t := time.Now().UTC()
        return &t
    })
}).
WithQuery(func(m *User, h query.PersistentHelper[User]) {
    h.Where().Field(&m.DeletedAt).EQ(nil)
})
```

Three required pieces:

1. **Marker column** (`DeletedAt`). Typically nullable — `*time.Time`. Add `OmitOnInsert` so it can't be accidentally written at INSERT time.
2. **`WithSoftDeletion`** — describes the value to write on "delete". The function runs on every `Delete` call and receives the context (useful for user/clock/tenant).
3. **`WithQuery` with a filter** — so soft-deleted records don't leak into SELECTs. Without it they show up in listings.

!!! note "SetValueFn return type"
    The returned value must match the field type — for `*time.Time` return `*time.Time`, not `time.Time`. `Build()` runs a type probe: each `SetValueFn` is invoked once with `context.Background()` and the returned value is checked against the field type. A mismatch (or a panic from inside the callback) is reported from `Build()` rather than crashing on the first soft `Delete()` call.

    This is not a full compile-time check — the callback body still runs at repo-build time, and if it branches on ctx values (`ctx.Value(tenantKey)`) that `context.Background()` does not carry, the probe exercises a different path than production. Keep `SetValueFn` bodies free of ctx-dependent branches when possible, or accept that the probe only catches the common ctx=Background case.

## How it works

`repo.Delete(ctx, …)` executes

```sql
UPDATE users SET deleted_at = ? WHERE …
```

instead of `DELETE FROM users WHERE …`. It returns the UPDATE's `RowsAffected`. If zero rows match, it returns `ErrNotFound`.

## Restoration

There's no dedicated API — restore a row with a direct UPDATE bypassing the repo:

```sql
UPDATE users SET deleted_at = NULL WHERE id = ?;
```

Alternatively, you can run an extra `repo.Update` with `Only(&m.DeletedAt)` and a `nil` value in the model — but that bypasses the persistent WHERE, so it only works when the repo's structure allows it.

## Multiple marker fields

`SoftDeletionBuilder` supports multiple `Field` calls — all of them will be updated on soft-delete:

```go
WithSoftDeletion(func(m *User, b *gerpo.SoftDeletionBuilder[User]) {
    b.Field(&m.DeletedAt).SetValueFn(now)
    b.Field(&m.DeletedBy).SetValueFn(userFromCtx)
})
```
