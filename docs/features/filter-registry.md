# Filter registry

The mapping **Go type → available WHERE operators** lives in a single global object: `filters.Registry`. Built-in buckets cover the types gerpo ships out of the box; the same registry is the extension point for your own types and for overriding default SQL.

## What the registry decides

Whenever a column is built (`Field(&m.X)` for plain columns, `WithCompute(...)` for non-aggregate virtual columns), gerpo asks the registry: *"for this Go type and this column SQL, which operators are valid and what SQL fragment does each operator emit?"* The result is stored on the column; runtime `WHERE` calls read from there.

A field with no registered operators still compiles into a repository — it just rejects `Where(...)` at request time with `option is not available`. The registry only enables operators; it never refuses to build a repository.

## Built-in buckets

| Bucket | Matches | Default operators |
|---|---|---|
| `filters.Registry.Bool` | `reflect.Kind == Bool` | `EQ`, `NotEQ` |
| `filters.Registry.String` | `reflect.Kind == String` | `EQ`, `NotEQ`, `In`, `NotIn`, the six `Contains/StartsWith/EndsWith` operators, plus every `*Fold` (case-insensitive) variant |
| `filters.Registry.Numeric` | every signed/unsigned int kind, `float32`, `float64` | `EQ`, `NotEQ`, `LT`, `LTE`, `GT`, `GTE`, `In`, `NotIn` |
| `filters.Registry.Time` | exact `reflect.Type == time.Time{}` | `LT`, `LTE`, `GT`, `GTE` |
| `filters.Registry.UUID` | exact `reflect.Type == uuid.UUID{}` | `EQ`, `NotEQ`, `In`, `NotIn` |

Pointer-wrapped fields (`*string`, `*time.Time`, …) additionally pick up `EQ` / `NotEQ` so `IS NULL` / `IS NOT NULL` work even on buckets that omit equality (`Time` for example).

`filters.Registry.Lookup(reflect.TypeOf(v))` returns the bucket registered for a custom type, or `nil`. `bucket.Operations()` returns the current operator list — handy for assertions in tests.

## Adding a custom type

A custom value type — `decimal.Decimal`, `civil.Date`, a `Money` struct — needs an explicit registration. Do it once at process start (typically in `init()` of the package that owns the type) so it is in place before any repository is built.

```go
package money

import (
    "github.com/insei/gerpo/filters"
    "github.com/insei/gerpo/types"
)

type Money struct {
    Amount   int64
    Currency string
}

func init() {
    filters.Registry.Register(Money{}).
        Allow(types.OperationEQ, types.OperationNotEQ,
            types.OperationLT, types.OperationLTE,
            types.OperationGT, types.OperationGTE,
            types.OperationIn, types.OperationNotIn)
}
```

`Allow(ops...)` wires each operator to the **stock SQL template** (`= ?`, `< ?`, `IN (?, ?, …)`, etc.). The template is the same one the built-in buckets use; you do not have to write SQL by hand for the common case.

`example` in `Register(example)` is read only for `reflect.TypeOf` — pass a zero value. Pointer wrappers (`*Money`) resolve to the same bucket via the dereference step, no separate registration needed.

## String aliases

A `type Status string` field falls through to the `String` bucket by default and inherits the full string operator set. That matches the historical behavior — but it ALSO trips a runtime guard if the user passes a literal `"active"`: the legacy filter rejected values whose `reflect.Type` did not exactly match the field type.

If you only need a narrow set, register the alias explicitly. The registry path uses a permissive equality check, so both `Status("active")` and the literal `"active"` work in `Where(...)`:

```go
type Status string

const (
    StatusActive   Status = "active"
    StatusArchived Status = "archived"
)

func init() {
    filters.Registry.Register(Status("")).
        Allow(types.OperationEQ, types.OperationIn)
}
```

```go
// Both calls succeed:
h.Where().Field(&m.Status).EQ(StatusActive)
h.Where().Field(&m.Status).EQ("active")
```

## Overriding a default

Use `Override(op, spec)` to replace the SQL fragment for one operator on any bucket — built-in or custom. The other operators on the bucket keep their defaults.

```go
// Compare timestamps by date only:
filters.Registry.Time.Override(types.OperationEQ, filters.Bound{
    SQL: "DATE_TRUNC('day', created_at) = DATE_TRUNC('day', CAST(? AS timestamptz))",
})
```

`Override` implicitly allows the operator if it was not in the bucket already — you do not have to call `Allow` first.

## Removing an operator

`Remove(ops...)` drops both the override (if any) and the default. Useful when a default does not make sense for your data:

```go
// Forbid LT on Age — domain rule says age comparisons go through a helper.
filters.Registry.Numeric.Remove(types.OperationLT)
```

## FilterSpec variants

`Override` takes a `filters.FilterSpec`. Five concrete shapes cover every common pattern:

| Variant | Use when |
|---|---|
| `filters.SQL("…")` | The fragment is fixed and binds no value. Example: `EXISTS (SELECT 1 …)`. |
| `filters.Bound{SQL: "…"}` | The fragment contains exactly one `?` and the user value is bound there. |
| `filters.SQLArgs{SQL: "…", Args: []any{…}}` | Predicate references constants the column owner already has; user value is **not** bound. |
| `filters.Match{Cases: …, Default: …}` | Discriminate on the user value (e.g. `true` → SQL A, `false` → SQL B) via `reflect.DeepEqual`. |
| `filters.Func(func(ctx, v) (sql, args, err) { … })` | Escape hatch for ctx-aware logic — multi-tenant, dynamic predicates. |

```go
filters.Registry.Bool.Override(types.OperationEQ, filters.Match{
    Cases: []filters.MatchCase{
        {Value: true,  Spec: filters.SQL("EXISTS (SELECT 1 FROM permissions WHERE …)")},
        {Value: false, Spec: filters.SQL("NOT EXISTS (SELECT 1 FROM permissions WHERE …)")},
    },
})
```

`virtual.WithFilter` consumes the same `FilterSpec` types — the `virtual.SQL` / `virtual.Bound` / … names are aliases for `filters.SQL` / `filters.Bound` / …, so existing virtual-column code keeps compiling unchanged.

## Resolution order

When the registry is consulted for a field, it walks this list in order and stops at the first match:

1. **Pointer fields** receive stock `EQ` / `NotEQ` first (for `IS NULL` semantics), then dereference.
2. A **custom-registered `reflect.Type`** wins over any kind bucket.
3. Named buckets `Time` and `UUID` match by exact `reflect.Type`.
4. Primitive **kind buckets** — `Bool`, `String`, `Numeric` — match by `reflect.Kind`.
5. **Unknown types** return an empty operator set. The repository still builds; runtime `Where(...)` on the column fails with `option is not available`.

## Tests that mutate the registry

`filters.Registry` is a single global. Tests that call `Register` / `Override` / `Remove` must restore the previous state, otherwise neighbouring tests inherit the change. `filters.Snapshot()` returns a `restore` function — pair it with `t.Cleanup`:

```go
func TestSomething(t *testing.T) {
    restore := filters.Snapshot()
    t.Cleanup(restore)

    filters.Registry.Register(Money{}).Allow(types.OperationEQ)
    // …test logic that assumes Money is registered…
}
```

`Snapshot` deep-copies the entire registry (built-in buckets and the custom-type map), so any combination of mutations rolls back atomically.

## Where to register

| Where | When |
|---|---|
| `init()` of the package that owns the type | Default. The type's filter contract sits next to its definition; every importer gets it transparently. |
| Top of `main()` | Use when the registration depends on configuration available only at runtime. |
| Inside a test, with `Snapshot` | When a test exercises a registration scenario without globally affecting other tests. |

Avoid registering inside repository factories — by the time the factory runs, the columns are already being built and stale registrations would silently miss.

## Limitations

- The registry is process-global. Two repositories in the same binary cannot disagree on the operator set for a given Go type.
- `Register` / `Override` / `Remove` are safe to call concurrently (each bucket holds an internal `sync.RWMutex`), but they should not race with column construction. Do registry mutations before `gerpo.New[T]().…Build()` is called.
- The registry does not validate at `Build()` time that a column has any operators. A typo in the type or a missed `init()` surfaces only when a `Where(...)` call hits the missing operator. If you want the strict check, walk the columns yourself: `column.GetAvailableFilterOperations()` returns the operator list per column.
