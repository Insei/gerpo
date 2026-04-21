# Persistent queries

`WithQuery(func(m *T, h query.PersistentHelper[T]))` defines conditions that apply to **every** request the repository runs — SELECT, COUNT, UPDATE, DELETE. Typical uses are soft delete, JOINs for virtual columns, GROUP BY.

## Four capabilities of PersistentHelper

| Method | Effect |
|---|---|
| `Where()` | Filters inserted into every query |
| `LeftJoinOn(table, on, resolver?)` / `InnerJoinOn(...)` | Static or per-request parameter-bound JOINs |
| `GroupBy(fields...)` | Override the auto GROUP BY (which kicks in for any aggregate virtual column) |
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
    c.Field(&m.PostCount).AsVirtual().Aggregate().Compute("COALESCE(COUNT(posts.id), 0)")
}).
WithQuery(func(m *User, h query.PersistentHelper[User]) {
    h.LeftJoinOn("posts", "posts.user_id = users.id")
    h.Where().Field(&m.DeletedAt).EQ(nil)
})
```

Now `PostCount` is automatically included in the SELECT of every request against `users`.

!!! tip "Auto GROUP BY"
    When at least one virtual column is marked with `.Aggregate()`, gerpo auto-fills GROUP BY with every non-aggregate column in SELECT. There is no more manual `h.GroupBy(...)` per repository — the type system already knows which columns are aggregates and which are not.

    Manual `h.GroupBy(...)` still works and takes precedence — power users can override the auto choice when the default doesn't fit (HAVING constructs, GROUP BY of expressions that are not in SELECT, ROLLUP, ...).

!!! note "InnerJoin vs LeftJoin"
    `InnerJoinOn` drops users who have no posts — handy when you only care about active ones. `LeftJoinOn` keeps them, the aggregate returns `0` for loners.

## Bound JOIN parameters

When the ON-clause needs runtime values (tenant id, locale, …), supply a resolver as the third argument. The persistent repository is built once, but the resolver runs **per request** and receives that request's `ctx`:

```go
h.LeftJoinOn(
    "posts",
    "posts.user_id = users.id AND posts.tenant_id = ?",
    func(ctx context.Context) ([]any, error) {
        tenantID, ok := ctx.Value(tenantKey{}).(string)
        if !ok {
            return nil, errors.New("tenant missing in ctx")
        }
        return []any{tenantID}, nil
    },
)
```

The values returned by the resolver flow through the driver's parameter binding in the order of the `?` placeholders, so they cannot turn into SQL — even if they originated in user input. A non-nil error from the resolver aborts the query: it is wrapped in `query.ErrApplyJoinClause` and surfaced as the `GetList`/`GetFirst`/… return value; no SQL is sent to the database.

Omit the resolver for a static JOIN:

```go
h.LeftJoinOn("posts", "posts.user_id = users.id")
```

`LeftJoinOn` / `InnerJoinOn` accept **at most one** resolver — passing two panics at registration time. The SQL template itself is always frozen at `WithQuery` time; the resolver only materialises the values for `?`. A raw-string-callback variant that lets ctx build SQL is deliberately absent — it was an SQL-injection hazard. Everything ctx-dependent must go through the resolver or through a matching per-request WHERE.

## Combining with per-request WHERE

Persistent conditions are joined with per-request conditions via AND and **always** come first. Your filter cannot disable a persistent WHERE, but it can add extra conditions.
