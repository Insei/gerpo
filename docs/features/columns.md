# Columns

Columns are described inside `Columns(func(m *T, c *gerpo.ColumnBuilder[T]))`. The binding is done **through a pointer to the struct field** — not via a string, not via a tag. The benefit: renaming a field is a regular refactor, and the compiler catches every mistake.

## Two kinds of columns

```go
c.Field(&m.Name)                 // regular, read and written — the default
c.Field(&m.FullName).AsVirtual().Compute("first_name || ' ' || last_name") // computed, read-only
```

### Regular column (default)

`Field(ptr)` registers a regular column right away — no extra step needed. The default SQL column name is `snake_case(field)` (e.g. `CreatedAt → created_at`); override with `WithColumnName`. All column-shaping methods (`OmitOnUpdate`, `WithAlias`, ...) are callable directly on the result of `Field`.

### Virtual column

Call `.AsVirtual()` to switch the field into a virtual (computed) column. Virtual columns are read-only — gerpo automatically protects them from both INSERT and UPDATE. Continue the chain with `.Compute(sql, args...)` (or `.Aggregate()` / `.Filter(op, spec)` for advanced cases). See [Virtual columns](virtual-columns.md).

## Options on regular columns

| Option | Effect |
|---|---|
| `WithColumnName(string)` | SQL column name (defaults to snake_case of the field name) |
| `WithTable(string)` | Table name — useful for columns coming from a JOIN |
| `WithAlias(string)` | Alias in SELECT |
| `OmitOnInsert()` | Exclude from INSERT (e.g. `UpdatedAt` set on UPDATE by a trigger/hook) |
| `OmitOnUpdate()` | Exclude from UPDATE SET (e.g. `CreatedAt` — set once, never changes) |
| `ReadOnly()` | Shortcut for `OmitOnInsert().OmitOnUpdate()` — SELECT-only |
| `ReturnedOnInsert()` | Include in `INSERT … RETURNING …`; the value is scanned back into the model field |
| `ReturnedOnUpdate()` | Include in `UPDATE … RETURNING …`; the value is scanned back into the model field |

## Common patterns

### PK with a database-side DEFAULT

```go
c.Field(&m.ID).ReadOnly()
```

`ID` appears only in WHERE and SELECT; it is never in INSERT, so the database generates it, and never in UPDATE, so it cannot be moved.

### created_at / updated_at

```go
c.Field(&m.CreatedAt).OmitOnUpdate()  // inserted, never updated
c.Field(&m.UpdatedAt).OmitOnInsert()  // set by a trigger or BeforeUpdate hook on UPDATE
```

### Server-generated values: RETURNING

Some columns get filled by the database — UUID PKs with `DEFAULT gen_random_uuid()`,
timestamps with `DEFAULT NOW()`, version counters bumped by a trigger. Mark them
with `ReturnedOnInsert()` / `ReturnedOnUpdate()` and gerpo emits a
`RETURNING …` clause and scans the value back into the model field, in-place.

```go
c.Field(&m.ID).ReadOnly().ReturnedOnInsert()         // PK with DEFAULT gen_random_uuid()
c.Field(&m.CreatedAt).ReadOnly().ReturnedOnInsert()  // DB DEFAULT NOW()
c.Field(&m.UpdatedAt).OmitOnInsert().ReturnedOnUpdate() // trigger-managed
c.Field(&m.Version).ReturnedOnInsert().ReturnedOnUpdate() // optimistic-lock counter
```

After `repo.Insert(ctx, &m)` the model carries the freshly generated ID,
CreatedAt, etc. After `repo.Update(ctx, &m, …)` the model carries the
trigger-bumped UpdatedAt / Version. There is no second SELECT round-trip and
no UUID generation on the application side.

For per-request control see [Insert](crud.md#insert) / [Update](crud.md#update)
helpers — both expose `Returning(fields...)` to narrow or disable RETURNING for
one call.

!!! note "Database support"
    `RETURNING` is a PostgreSQL-style feature; SQLite ≥3.35 and MariaDB ≥10.5
    also support it. The bundled `pgx5`, `pgx4` and `databasesql` adapters all
    target PostgreSQL-compatible databases. MySQL has no `RETURNING` — see
    [TODO.md](https://github.com/insei/gerpo/blob/main/TODO.md) for the plan
    around multi-dialect support.

### Column from a JOIN

```go
c.Field(&m.PostTitle).WithTable("posts")
```

SELECT will read `posts.post_title`. The JOIN itself must be configured via [Persistent queries](persistent-queries.md).

### Nullable columns

Use a pointer type — `*string`, `*time.Time`, `*bool`. The `EQ(nil)` / `NotEQ(nil)` operators work and gerpo generates `IS NULL` / `IS NOT NULL`.

```go
type User struct {
    Email *string // nullable email
}
c.Field(&m.Email)

// Query:
repo.Count(ctx, func(m *User, h query.CountHelper[User]) {
    h.Where().Field(&m.Email).EQ(nil) // WHERE email IS NULL
})
```

## Columns storage and ExecutionColumns

`repo.GetColumns()` returns `types.ColumnsStorage` — the collection of every configured `Column`. On each request, gerpo creates an `ExecutionColumns` — a slice filtered by the specific action (SELECT / INSERT / UPDATE / …) taking protections and `Exclude/Only` helpers into account.

Direct access to `ColumnsStorage` is useful if you need to build a custom query and bypass the repo — rarely needed in practice.
