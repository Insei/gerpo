# Persistent queries

`WithQuery(func(m *T, h query.PersistentHelper[T]))` defines conditions that apply to **every** request the repository runs — SELECT, COUNT, UPDATE, DELETE. Typical uses are soft delete, JOINs for virtual columns, GROUP BY.

## Four capabilities of PersistentHelper

| Method | Effect |
|---|---|
| `Where()` | Filters inserted into every query |
| `LeftJoin(fn)` / `InnerJoin(fn)` | JOINs — body is returned by a `func(context.Context) string` |
| `GroupBy(fields...)` | A single GROUP BY applied everywhere (required when a JOIN + aggregate shows up) |
| `Exclude(fields...)` | Hide a column from every SELECT |

## Hiding soft-deleted records

```go
.WithQuery(func(m *User, h query.PersistentHelper[User]) {
    h.Where().Field(&m.DeletedAt).EQ(nil)
})
```

Now every `GetFirst`/`GetList`/`Count` automatically ignores records whose `DeletedAt` is non-null.

## JOIN + virtual column

A real example from the integration tests — `User` has a virtual `PostCount` field computed through a LEFT JOIN on `posts`:

```go
.Columns(func(m *User, c *gerpo.ColumnBuilder[User]) {
    c.Field(&m.ID)
    c.Field(&m.Name)
    c.Field(&m.PostCount).AsVirtual().WithSQL(func(ctx context.Context) string {
        return "COALESCE(COUNT(posts.id), 0)"
    })
}).
WithQuery(func(m *User, h query.PersistentHelper[User]) {
    h.LeftJoin(func(ctx context.Context) string {
        return "posts ON posts.user_id = users.id"
    })
    h.GroupBy(&m.ID, &m.Name)
    h.Where().Field(&m.DeletedAt).EQ(nil)
})
```

Now `PostCount` is automatically included in the SELECT of every request against `users`.

!!! note "InnerJoin vs LeftJoin"
    `InnerJoin` drops users who have no posts — handy when you only care about active ones. `LeftJoin` keeps them, the aggregate returns `0` for loners.

## Bound JOIN parameters — `LeftJoinOn` / `InnerJoinOn`

When the ON-clause needs runtime values (tenant id, locale, …), use the
parameter-bound forms. They take the joined table reference, the ON clause
with `?` placeholders, and bound arguments — exactly like a WHERE.

```go
h.LeftJoinOn(
    "posts",
    "posts.user_id = users.id AND posts.tenant_id = ?",
    tenantID,
)
```

The arguments flow through the driver's parameter binding, so values cannot
turn into SQL — even if `tenantID` originated in user input.

`InnerJoinOn` works the same way for inner joins.

## Legacy callback JOIN (deprecated)

The original `LeftJoin(fn)` / `InnerJoin(fn)` helpers take a callback that
returns the JOIN body. The callback receives a `context.Context`, but the
returned string is inlined verbatim — values are NOT parameterised:

```go
h.LeftJoin(func(ctx context.Context) string {
    tenantID := ctxpkg.TenantID(ctx)
    return fmt.Sprintf(
        "posts ON posts.user_id = users.id AND posts.tenant_id = '%s'",
        tenantID,
    )
})
```

!!! danger "SQL injection"
    Anything you splice into the returned string lands in the SQL as text.
    The callback form remains for backwards compatibility but is **deprecated**;
    new code should use `LeftJoinOn` / `InnerJoinOn`.

## Combining with per-request WHERE

Persistent conditions are joined with per-request conditions via AND and **always** come first. Your filter cannot disable a persistent WHERE, but it can add extra conditions.
