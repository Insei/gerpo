# Field mapping

gerpo's pointer-based column binding is powered by [`github.com/insei/fmap/v3`](https://github.com/insei/fmap), a small library that learns the layout of a struct once and exposes O(1) access to each field through its pointer address.

## Why not reflect.Type?

Classic reflection identifies a field by name or index. Both are brittle: renames silently shift indices, typos escape the compiler. gerpo wants code like

```go
c.Field(&m.Email)
```

where `&m.Email` is the reference. That needs a trick: convert a pointer-to-field into a stable field identifier.

## How fmap does it

1. On `gerpo.NewBuilder[T]()`, gerpo creates a zero `T` on the heap once.
2. `fmap` walks the struct using reflection and records each field's `unsafe.Offsetof`.
3. For every `c.Field(&m.X)` the builder takes the pointer `&m.X`, subtracts the base address of the zero-struct, and looks up the field by offset. O(1).
4. At query time, when the repo has a real `*T`, it reads/writes the field by adding the offset back — `unsafe.Pointer` arithmetic without `reflect.Value`.

Pointer-to-field resolution happens once per repo build, not per request. Per-request, gerpo only does pointer arithmetic and interface boxing — no `reflect.Value` on hot paths.

## What fmap returns

`fmap.Field` implements enough to:

- fetch the Go type (`GetType`, `GetDereferencedType`) — used when generating operators for a given column;
- read the value (`Get(model)`) — used by `GetModelValues` when building INSERT/UPDATE;
- write the value (`Set(model, val)`) — used by `SoftDeletionBuilder.SetValueFn`;
- obtain a pointer (`GetPtr(model)`) — used by `GetModelPointers` for `rows.Scan`.

`column.Column` wraps an `fmap.Field` plus the operator table, table name, alias, and allowed SQL actions.

## Allocation behaviour

fmap allocates once per struct (not per field, not per query). Its `Set` uses a compiled set of pointer-tricks per Go kind — no `reflect.Call`, no boxing on the hot path.

gerpo's own overhead on top is:

- one `[]any` for `GetModelPointers` (required by `rows.Scan`);
- one `[]any` for `GetModelValues` on INSERT/UPDATE (required by the driver to marshal parameters);
- one allocation per closure stored in the WHERE/ORDER plan.

Those slices are a known allocation source — see the backlog of allocation ideas in the repository's memory.

## Limitations

- Nested anonymous structs are supported; unexported fields are not.
- Interface-typed fields can be columns, but conversion to SQL values depends on the driver. Prefer concrete types.
- `fmap` uses `unsafe`, so the usual caveats apply: struct layout must match what `fmap` saw at init, i.e. no hot-swapping types or binary-incompatible upgrades.
