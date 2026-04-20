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
