# Virtual columns

Virtual columns are SELECT expressions mapped onto struct fields. They are **read-only**: automatically protected from both INSERT and UPDATE.

## Compute — the simple case

```go
type User struct {
    FirstName string
    LastName  string
    FullName  string // virtual
}

c.Field(&m.FullName).AsVirtual().Compute("first_name || ' ' || last_name")
```

Compute wraps the expression in parentheses automatically — `(first_name || ' ' || last_name)` — so it composes cleanly inside larger predicates. This is part of the contract, not magic: you can rely on it.

SELECT receives the parenthesised expression (no alias — aliases on virtual columns are not emitted today), and the Scan drops the value into `m.FullName`.

## Compute with bound args

The second form binds positional parameters into the expression itself — useful when the SQL depends on a constant you want the driver to parameterise, rather than Sprintf (and the SQL-injection risk that carries).

```go
c.Field(&m.PostCount).AsVirtual().Compute(
    "SELECT count(*) FROM posts WHERE posts.user_id = users.id AND posts.title LIKE ?",
    "%post%",
)
```

Wherever the column is referenced — SELECT, auto-derived WHERE filter, ORDER BY — gerpo injects the args into the final bound-args list in the correct position.

## Auto-derived filters

Compute-built virtual columns inherit the full set of operators that a regular column of the same Go type would support. For example, a `string`-typed virtual column gets `EQ`, `NEQ`, `IN`, `NIN`, `CT`, `CT_IC`, and the other LIKE variants for free. For numeric types you also get `LT`/`LTE`/`GT`/`GTE`.

The predicate takes the form `(compute_sql) op ?`, so the column's bound args come first, followed by the user value.

If you don't want that (aggregates, ctx-aware SQL), see [Aggregate](#aggregate) and [Filter](#filter-escape-hatch) below.

## Aggregations from a JOIN

Virtual columns often aggregate related tables. You need a pair: a JOIN in the persistent query, and a matching GROUP BY.

```go
c.Field(&m.PostCount).AsVirtual().Compute("COALESCE(COUNT(posts.id), 0)")

.WithQuery(func(m *User, h query.PersistentHelper[User]) {
    h.LeftJoin(func(ctx context.Context) string {
        return "posts ON posts.user_id = users.id"
    })
    h.GroupBy(&m.ID, &m.Name /*, every non-aggregate */)
})
```

!!! warning "GROUP BY"
    When aggregates are in play, you must list every regular column that appears in SELECT in the `GroupBy`, otherwise PostgreSQL returns *"must appear in the GROUP BY clause or be used in an aggregate function"*.

## Aggregate

`Aggregate()` marks a column as an aggregate expression. The only practical effect today: **filtering on an aggregate column without an explicit `Filter()` override is rejected by the WhereBuilder** with a clear error, instead of producing invalid SQL (`COUNT(...)` inside a WHERE clause). There is no auto-routing to HAVING — if you need HAVING semantics, register a `Filter` that expands the condition the way your dialect expects.

```go
c.Field(&m.PostCount).AsVirtual().
    Aggregate().
    Compute("COALESCE(COUNT(posts.id), 0)")
```

## Filter (escape hatch)

`Filter(op, spec)` overrides the SQL used for a single operator. Other operators keep their auto-derived implementations — this is contractual. Aggregate columns have no auto-derived operators at all, so every operator you want available has to be registered through `Filter`.

`spec` is a `FilterSpec` — one of five variants covering the realistic patterns:

### virtual.SQL — static fragment

```go
.Filter(types.OperationEQ, virtual.SQL("EXISTS (SELECT 1 FROM sessions WHERE user_id = users.id)"))
```

The user value is not appended to bound args.

### virtual.Bound — fragment with one placeholder

```go
.Filter(types.OperationGT, virtual.Bound{SQL: "SUM(amount) > ?"})
```

Exactly one `?` in the SQL receives the user value.

### virtual.SQLArgs — fragment with explicit bound args

```go
.Filter(types.OperationEQ, virtual.SQLArgs{
    SQL:  "computed_at BETWEEN ? AND ?",
    Args: []any{from, to},
})
```

The user value is **not** bound — the override is self-contained. `Args` is copied defensively at registration time so you can safely mutate the slice afterwards.

### virtual.Match — discriminate on the user value

```go
.Filter(types.OperationEQ, virtual.Match{
    Cases: []virtual.MatchCase{
        {Value: true,  Spec: virtual.SQL("EXISTS (SELECT 1 FROM tokens WHERE user_id = users.id)")},
        {Value: false, Spec: virtual.SQL("NOT EXISTS (SELECT 1 FROM tokens WHERE user_id = users.id)")},
    },
    Default: virtual.SQL("FALSE"),
})
```

The first `Case` whose `Value` is equal (per `reflect.DeepEqual`) to the user value wins. `Default` is optional — a nil `Default` with no match yields a clear error from `Apply`. Cases can be any FilterSpec, including another `Match`.

### virtual.Func — escape hatch for ctx-aware SQL

When the SQL genuinely has to look at `context.Context` — multi-tenant, dynamic shard, audit tracing — use `Func`:

```go
.Filter(types.OperationEQ, virtual.Func(func(ctx context.Context, v any) (string, []any, error) {
    tid := ctx.Value(tenantKey).(int)
    return "x.tenant_id = ? AND x.flag = ?", []any{tid, v}, nil
}))
```

`Func` is the only variant that can see `ctx`. Everything else is static and can be tested by comparing structs.

## Read-only in practice

Trying to assign a value to a virtual field during `Insert`/`Update` isn't an error — gerpo simply ignores it:

```go
u := &User{PostCount: 9999} // never stored
repo.Insert(ctx, u)
```

The next `GetFirst` returns whatever the database computed.

## Migration: deprecated → new

| Deprecated                                   | New                                                              |
|----------------------------------------------|------------------------------------------------------------------|
| `WithSQL(func(ctx) string)`                  | `Compute(sql string, args ...any)`                               |
| `WithBoolEqFilter(fn)`                       | `Filter(types.OperationEQ, virtual.Match{Cases: ..., Default})`  |

`WithSQL` and `WithBoolEqFilter` still compile; the deprecation bracket is 1–2 minor releases. Existing code keeps working — plan the migration before the next major bump.
