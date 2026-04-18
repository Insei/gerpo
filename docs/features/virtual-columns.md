# Virtual columns

Virtual columns are SELECT expressions mapped onto struct fields. They are **read-only**: automatically protected from both INSERT and UPDATE.

## Simple computed field

```go
type User struct {
    FirstName string
    LastName  string
    FullName  string // virtual
}

c.Field(&m.FullName).AsVirtual().WithSQL(func(ctx context.Context) string {
    return "first_name || ' ' || last_name"
})
```

SELECT will get `first_name || ' ' || last_name AS full_name` (alias is the snake_case of the field), and the Scan drops the value into `m.FullName`.

## Aggregations from a JOIN

Virtual columns often aggregate related tables. You need a pair: a JOIN in the persistent query, and a matching GROUP BY.

```go
c.Field(&m.PostCount).AsVirtual().WithSQL(func(ctx context.Context) string {
    return "COALESCE(COUNT(posts.id), 0)"
})

.WithQuery(func(m *User, h query.PersistentHelper[User]) {
    h.LeftJoin(func(ctx context.Context) string {
        return "posts ON posts.user_id = users.id"
    })
    h.GroupBy(&m.ID, &m.Name /*, every non-aggregate */)
})
```

!!! warning "GROUP BY"
    When aggregates are in play, you must list every regular column that appears in SELECT in the `GroupBy`, otherwise PostgreSQL returns *"must appear in the GROUP BY clause or be used in an aggregate function"*.

## Context-aware SQL

`WithSQL(fn)` takes a `context.Context`, so the SQL is produced per request. You can mix in data from the context (tenant, UI locale), but **values are not parameterized** — be careful with user input.

## Read-only in practice

Trying to assign a value to a virtual field during `Insert`/`Update` isn't an error — gerpo simply ignores it:

```go
u := &User{FullName: "fake"} // never stored
repo.Insert(ctx, u)
```

The next `GetFirst` returns whatever the database computed.

## Boolean virtuals with custom filter (deprecated API)

For `bool`-typed virtual columns that need different SQL for `true/false/nil` in WHERE, there is `WithBoolEqFilter`:

```go
c.Field(&m.IsActive).AsVirtual().
    WithSQL(func(ctx context.Context) string { return "EXISTS (SELECT 1 FROM sessions WHERE user_id = users.id)" }).
    WithBoolEqFilter(func(b *virtual.BoolEQFilterBuilder) {
        b.AddTrueSQLFn(func(ctx context.Context) string { return "EXISTS (...)" })
        b.AddFalseSQLFn(func(ctx context.Context) string { return "NOT EXISTS (...)" })
    })
```

!!! note "Deprecated"
    The current virtual-filter API is marked as deprecated in the 1.0.0 roadmap — a new configuration API for virtual columns is planned. It still works today; just be ready to migrate.
