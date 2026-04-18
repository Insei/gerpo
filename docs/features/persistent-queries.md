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
    c.Field(&m.ID).AsColumn()
    c.Field(&m.Name).AsColumn()
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

## Context-aware JOIN

The function returns the JOIN text and takes a `context.Context`. That lets you mix in runtime values (tenant ID, UI locale) into the body:

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
    Values coming from the context into the JOIN body do not flow through parameter binding. If you interpolate user-supplied data, escape it yourself — or, better, switch to a WHERE with a bound parameter.

## Combining with per-request WHERE

Persistent conditions are joined with per-request conditions via AND and **always** come first. Your filter cannot disable a persistent WHERE, but it can add extra conditions.
