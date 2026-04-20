# Hooks

Hooks are functions that gerpo runs around repository operations. Useful for generated fields, auditing, projections, and user-land cascade writes.

## Hook kinds

| Option | Signature | Called |
|---|---|---|
| `WithBeforeInsert` | `func(ctx, *T) error` | before SQL `INSERT` |
| `WithAfterInsert` | `func(ctx, *T) error` | after a successful `INSERT` |
| `WithBeforeInsertMany` | `func(ctx, []*T) error` | before the batched `INSERT` (`InsertMany`) |
| `WithAfterInsertMany` | `func(ctx, []*T) error` | after a successful `InsertMany` |
| `WithBeforeUpdate` | `func(ctx, *T) error` | before SQL `UPDATE` |
| `WithAfterUpdate` | `func(ctx, *T) error` | after a successful `UPDATE` (rowsAffected > 0) |
| `WithAfterSelect` | `func(ctx, []*T) error` | after Scan of `GetFirst`/`GetList` |

For `GetFirst`, `afterSelect` receives a single-element slice. For `GetList` — the full slice.

`InsertMany` has its **own** pair of hooks (`…InsertMany`), not the single-row ones. The single-row `WithBeforeInsert` / `WithAfterInsert` do **not** fire per row when you call `InsertMany`. Separate hooks let the cascade case — "I inserted N parents, now write N-ish children" — issue one batched child query instead of N serial ones. See [Cascading related rows](#cascading-related-rows-user-land-one-to-many) below.

## Error contract

Every hook returns `error`. The rules are symmetric across all five:

- A **`Before*` hook** returning non-nil **aborts the operation**. The SQL does NOT run, and the error is returned to the caller (after passing through [WithErrorTransformer](error-transformer.md), if any).
- An **`After*` hook** returning non-nil **surfaces the error after the SQL already ran**. The row is already written / updated / fetched; gerpo does not roll anything back automatically. If the operation is inside a `gerpo.RunInTx`, the returned error is what `RunInTx` uses to decide between commit and rollback — so wrapping in a transaction is how you make an `After*` hook's failure undo the side effects.

## Mutating the model

Changes made in a `Before*` hook land in the SQL. This is the standard way to fill generated fields:

```go
.WithBeforeInsert(func(ctx context.Context, u *User) error {
    if u.ID == uuid.Nil {
        u.ID = uuid.New()
    }
    if u.CreatedAt.IsZero() {
        u.CreatedAt = time.Now().UTC()
    }
    return nil
}).
WithBeforeUpdate(func(ctx context.Context, u *User) error {
    now := time.Now().UTC()
    u.UpdatedAt = &now
    return nil
})
```

`After*` hooks can also mutate the model, but their changes never reach the database — they only affect the caller's copy.

## Stacking

`WithBeforeInsert` and its siblings can be called multiple times — hooks run in registration order, and the first non-nil error stops the chain:

```go
.WithBeforeInsert(setDefaults).
WithBeforeInsert(audit.LogAttempt)
```

## Cascading related rows (user-land one-to-many)

gerpo does not model relations itself (no lazy load, no `has_many`), but the combination of `AfterInsert`/`AfterUpdate`, the ctx-carried transaction from [`gerpo.WithTx` / `gerpo.RunInTx`](transactions.md) and the new hook error return gives you a clean, explicit mechanism to cascade related rows in the same transaction.

```go
orderRepo, _ := gerpo.New[Order]().
    Adapter(adapter).Table("orders").
    Columns(func(m *Order, c *gerpo.ColumnBuilder[Order]) {
        c.Field(&m.ID).OmitOnUpdate()
        c.Field(&m.UserID)
        c.Field(&m.Total)
    }).
    WithAfterInsert(func(ctx context.Context, o *Order) error {
        // o.Items is a []OrderItem that lives in a different table.
        for i := range o.Items {
            o.Items[i].OrderID = o.ID
            if err := itemRepo.Insert(ctx, &o.Items[i]); err != nil {
                return err // RunInTx will roll the whole thing back
            }
        }
        return nil
    }).Build()

// usage
err := gerpo.RunInTx(ctx, adapter, func(ctx context.Context) error {
    return orderRepo.Insert(ctx, &order)
    // AfterInsert fires → inserts items in the SAME transaction.
    // Any item failure → error propagates → RunInTx rolls the order back too.
})
```

Why this works:

- `gerpo.RunInTx` begins a transaction and puts it into `ctx` via `gerpo.WithTx`. Every Repository operation invoked with that ctx picks up the tx automatically — no arguments to thread through.
- The `itemRepo.Insert(ctx, …)` call inside the hook sees the same ctx and the same tx. The child row lands in the same transaction.
- Any error — from the cascade body, from the child Insert, from a deeper nested hook — bubbles up through `repo.Insert → RunInTx → your code`, and `RunInTx` rolls back on non-nil.

**Watch out for:** if the cascading hook itself triggers more hooks on other repos (an `AfterInsert` on item triggers another `AfterInsert` elsewhere), you end up with an implicit recursion tree. Keep the cascade graph explicit and shallow.

### Cascading a batched insert

`WithAfterInsertMany` receives the full parent slice, so you can collect child
rows across every parent and write them in **one** `itemRepo.InsertMany(ctx, …)`
call — avoiding the N+1 pattern that falls out naturally if you reuse the
single-row hook.

```go
orderRepo, _ := gerpo.New[Order]().
    Adapter(adapter).Table("orders").
    Columns(/* … */).
    WithAfterInsertMany(func(ctx context.Context, orders []*Order) error {
        // Collect all items across all parents.
        var items []*OrderItem
        for _, o := range orders {
            for i := range o.Items {
                o.Items[i].OrderID = o.ID
                items = append(items, &o.Items[i])
            }
        }
        if len(items) == 0 {
            return nil
        }
        _, err := itemRepo.InsertMany(ctx, items)
        return err // RunInTx rolls the whole batch back on non-nil
    }).Build()
```

## Typical uses

- **Field auto-fill:** IDs, timestamps, tenant_id.
- **Auditing:** log INSERT/UPDATE/DELETE together with the user and payload.
- **Cascade related rows:** see above.
- **Post-processing in `AfterSelect`:** e.g. decrypting encrypted fields.

## Validation is a poor fit

You *could* return an error from `BeforeInsert` to reject "invalid" inputs, but validation belongs at the service/domain layer — the Repository is the persistence layer and should trust what arrives. If you catch an invalid payload inside `BeforeInsert`, a corresponding caller has already slipped past a validation boundary higher up; fix that instead.

## What you can't do in a hook

- Open a **separate** transaction on the same connection — you are already inside a gerpo call. Use the ambient tx from ctx, or run external side effects outside `.Insert`/`.Update`.
- Catch a DB error from the SQL itself — by the time an `After*` hook runs, the SQL succeeded. Map raw DB errors via [Error transformer](error-transformer.md) instead.
