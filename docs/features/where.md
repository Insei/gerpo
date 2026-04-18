# WHERE operators

All operators are built with `h.Where().Field(&m.X).<Op>(val)`. The result is an `ANDOR` that lets you keep chaining with `.AND()` / `.OR()`.

## Comparison

| Method | SQL | Works for |
|---|---|---|
| `EQ(v)` | `= ?` (or `IS NULL` when `v == nil`) | any type |
| `NEQ(v)` | `!= ?` (or `IS NOT NULL`) | any |
| `LT(v)` | `< ?` | numbers, dates |
| `LTE(v)` | `<= ?` | numbers, dates |
| `GT(v)` | `> ?` | numbers, dates |
| `GTE(v)` | `>= ?` | numbers, dates |

```go
h.Where().Field(&m.Age).GTE(18)
h.Where().Field(&m.DeletedAt).EQ(nil) // IS NULL
```

## Sets

| Method | SQL |
|---|---|
| `IN(a, b, c)` | `IN (?, ?, ?)` |
| `NIN(a, b, c)` | `NOT IN (?, ?, ?)` |

Accept variadic `any` or an already-expanded slice:

```go
h.Where().Field(&m.ID).IN(id1, id2, id3)
h.Where().Field(&m.ID).IN(ids...) // if ids is []uuid.UUID
```

## String patterns

| Method | SQL | Case-insensitive variant |
|---|---|---|
| `CT(v)` | `LIKE CONCAT('%', CAST(? AS text), '%')` | `CT(v, true)` → `LOWER(col) LIKE LOWER(…)` |
| `NCT(v)` | `NOT LIKE CONCAT('%', …, '%')` | `NCT(v, true)` |
| `BW(v)` | `LIKE CONCAT(CAST(? AS text), '%')` | `BW(v, true)` |
| `NBW(v)` | `NOT LIKE CONCAT(…, '%')` | `NBW(v, true)` |
| `EW(v)` | `LIKE CONCAT('%', CAST(? AS text))` | `EW(v, true)` |
| `NEW(v)` | `NOT LIKE CONCAT('%', …)` | `NEW(v, true)` |

```go
h.Where().Field(&m.Title).CT("go", true)       // case-insensitive contains
h.Where().Field(&m.Email).BW("admin@")         // starts with
h.Where().Field(&m.Path).EW(".log", true)      // ends with, case-insensitive
```

!!! note "CAST(? AS text)"
    A bare `CONCAT(…)` breaks PostgreSQL type inference, so gerpo casts the parameter to `text` explicitly. The same form is valid in MySQL.

## Logic: AND, OR, Group

- Consecutive `Field(...).Op(...)` calls are joined by an **implicit AND**.
- `.OR()` / `.AND()` in the chain add explicit joiners.
- `Group(func(t WhereTarget))` produces parentheses.

```go
// (age >= 18 AND email IS NOT NULL) OR role = 'admin'
h.Where().Group(func(t types.WhereTarget) {
    t.Field(&m.Age).GTE(18).
        AND().Field(&m.Email).NEQ(nil)
}).OR().Field(&m.Role).EQ("admin")
```

## Custom operation — `OP`

If a column exposes a custom filter (relevant for [virtual columns](virtual-columns.md) with `WithBoolEqFilter`), invoke it by name:

```go
h.Where().Field(&m.IsActive).OP(types.OperationEQ, true)
```

## Limitations

- `LT`/`GT`/`LTE`/`GTE` do not type-check at runtime — the database does. gerpo passes values through as-is.
- For string LIKE operators the value must be a string; for `EQ`/`NEQ` the value type must match the field type (nullable types accept `nil`).
- A type mismatch produces a descriptive error wrapped with `gerpo.ErrApplyQuery`.
