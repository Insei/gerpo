# WHERE operators

All operators are built with `h.Where().Field(&m.X).<Op>(val)`. The result is an `ANDOR` that lets you keep chaining with `.AND()` / `.OR()`.

!!! tip "Static type-check at `go vet` time"
    Every operator on this page accepts `any`, so the compiler cannot catch `EQ("18")` on an `int` field. gerpo ships **[gerpolint](static-analysis.md)** — a `go/analysis` checker that flags these mismatches, either as a standalone binary or as a golangci-lint plugin.

## Comparison

| Method | SQL | Works for |
|---|---|---|
| `EQ(v)` | `= ?` (or `IS NULL` when `v == nil`) | any type |
| `NotEQ(v)` | `!= ?` (or `IS NOT NULL`) | any |
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
| `In(a, b, c)` | `IN (?, ?, ?)` |
| `NotIn(a, b, c)` | `NOT IN (?, ?, ?)` |

Accept variadic `any` or an already-expanded slice:

```go
h.Where().Field(&m.ID).In(id1, id2, id3)
h.Where().Field(&m.ID).In(ids...) // if ids is []uuid.UUID
```

## String patterns

String-typed (and `*string`) columns get six LIKE-style operators:

| Method | SQL |
|---|---|
| `Contains(v)` | `LIKE CONCAT('%', CAST(? AS text), '%')` |
| `NotContains(v)` | `NOT LIKE CONCAT('%', …, '%')` |
| `StartsWith(v)` | `LIKE CONCAT(CAST(? AS text), '%')` |
| `NotStartsWith(v)` | `NOT LIKE CONCAT(…, '%')` |
| `EndsWith(v)` | `LIKE CONCAT('%', CAST(? AS text))` |
| `NotEndsWith(v)` | `NOT LIKE CONCAT('%', …)` |

```go
h.Where().Field(&m.Title).Contains("go")
h.Where().Field(&m.Email).StartsWith("admin@")
h.Where().Field(&m.Path).EndsWith(".log")
```

!!! note "CAST(? AS text)"
    A bare `CONCAT(…)` breaks PostgreSQL type inference, so gerpo casts the parameter to `text` explicitly. The same form is valid in MySQL.

## Case-insensitive (`Fold`) variants

The `Fold` suffix is the Go-idiomatic spelling for case-insensitive equality (`strings.EqualFold`). All case-insensitive string operators follow the same naming:

| Method | SQL |
|---|---|
| `EQFold(v)` | `LOWER(col) = LOWER(CAST(? AS text))` |
| `NotEQFold(v)` | `LOWER(col) != LOWER(CAST(? AS text))` |
| `ContainsFold(v)` | `LOWER(col) LIKE LOWER(CONCAT('%', CAST(? AS text), '%'))` |
| `NotContainsFold(v)` | `LOWER(col) NOT LIKE LOWER(CONCAT('%', …, '%'))` |
| `StartsWithFold(v)` | `LOWER(col) LIKE LOWER(CONCAT(CAST(? AS text), '%'))` |
| `NotStartsWithFold(v)` | `LOWER(col) NOT LIKE LOWER(CONCAT(…, '%'))` |
| `EndsWithFold(v)` | `LOWER(col) LIKE LOWER(CONCAT('%', CAST(? AS text)))` |
| `NotEndsWithFold(v)` | `LOWER(col) NOT LIKE LOWER(CONCAT('%', …))` |

```go
h.Where().Field(&m.Email).EQFold("alice@Example.com")
h.Where().Field(&m.Title).ContainsFold("go")
h.Where().Field(&m.Path).EndsWithFold(".LOG")
```

All `*Fold` operators are registered only on string-typed columns (gerpo does not lowercase numbers or timestamps).

## Logic: AND, OR, Group

- Consecutive `Field(...).Op(...)` calls are joined by an **implicit AND**.
- `.OR()` / `.AND()` in the chain add explicit joiners.
- `Group(func(t WhereTarget))` produces parentheses.

```go
// (age >= 18 AND email IS NOT NULL) OR role = 'admin'
h.Where().Group(func(t types.WhereTarget) {
    t.Field(&m.Age).GTE(18).
        AND().Field(&m.Email).NotEQ(nil)
}).OR().Field(&m.Role).EQ("admin")
```

## Custom types and overrides

The list of operators each Go type accepts (and the SQL fragment each one emits) lives in `filters.Registry`. Adding `decimal.Decimal`, a string-alias, or a `Money` struct — and overriding the default SQL for `time.Time.EQ`, for example — happens through that registry. See [Filter registry](filter-registry.md).

## Limitations

- `LT`/`GT`/`LTE`/`GTE` do not type-check at runtime — the database does. gerpo passes values through as-is.
- For string LIKE operators the value must be a string; for `EQ`/`NotEQ` the value type must match the field type (nullable types accept `nil`). String aliases (`type Status string`) and other custom types relax this rule once registered — see [Filter registry](filter-registry.md).
- A type mismatch produces a descriptive error wrapped with `gerpo.ErrApplyQuery`.
