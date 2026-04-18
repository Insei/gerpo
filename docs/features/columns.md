# Columns

Columns are described inside `Columns(func(m *T, c *gerpo.ColumnBuilder[T]))`. The binding is done **through a pointer to the struct field** — not via a string, not via a tag. The benefit: renaming a field is a regular refactor, and the compiler catches every mistake.

## Two kinds of columns

```go
c.Field(&m.Name).AsColumn()      // regular, read and written
c.Field(&m.FullName).AsVirtual() // computed, read-only
    .WithSQL(func(ctx context.Context) string { return "first_name || ' ' || last_name" })
```

### `AsColumn` — regular column

Maps to a physical table column. The default column name is `snake_case(field)` (e.g. `CreatedAt → created_at`). Override with `WithColumnName`.

### `AsVirtual` — virtual column

Not stored in the table — computed by a SQL expression on the fly. Automatically protected from both INSERT and UPDATE. See [Virtual columns](virtual-columns.md).

## Options on regular columns

| Option | Effect |
|---|---|
| `WithColumnName(string)` | SQL column name (defaults to snake_case of the field name) |
| `WithTable(string)` | Table name — useful for columns coming from a JOIN |
| `WithAlias(string)` | Alias in SELECT |
| `WithInsertProtection()` | Exclude from INSERT (e.g. for a PK with DEFAULT) |
| `WithUpdateProtection()` | Exclude from UPDATE SET (e.g. for `created_at`) |

## Common patterns

### PK with a database-side DEFAULT

```go
c.Field(&m.ID).AsColumn().WithInsertProtection().WithUpdateProtection()
```

`ID` appears only in WHERE and SELECT; it is never in INSERT, so the database generates it.

### created_at / updated_at

```go
c.Field(&m.CreatedAt).AsColumn().WithUpdateProtection()  // inserted, never updated
c.Field(&m.UpdatedAt).AsColumn().WithInsertProtection()  // set by a trigger on UPDATE
```

### Column from a JOIN

```go
c.Field(&m.PostTitle).AsColumn().WithTable("posts")
```

SELECT will read `posts.post_title`. The JOIN itself must be configured via [Persistent queries](persistent-queries.md).

### Nullable columns

Use a pointer type — `*string`, `*time.Time`, `*bool`. The `EQ(nil)` / `NEQ(nil)` operators work and gerpo generates `IS NULL` / `IS NOT NULL`.

```go
type User struct {
    Email *string // nullable email
}
c.Field(&m.Email).AsColumn()

// Query:
repo.Count(ctx, func(m *User, h query.CountHelper[User]) {
    h.Where().Field(&m.Email).EQ(nil) // WHERE email IS NULL
})
```

## Columns storage and ExecutionColumns

`repo.GetColumns()` returns `types.ColumnsStorage` — the collection of every configured `Column`. On each request, gerpo creates an `ExecutionColumns` — a slice filtered by the specific action (SELECT / INSERT / UPDATE / …) taking protections and `Exclude/Only` helpers into account.

Direct access to `ColumnsStorage` is useful if you need to build a custom query and bypass the repo — rarely needed in practice.
