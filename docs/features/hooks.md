# Hooks

Hooks are functions that gerpo runs around repository operations. Useful for generated fields, auditing, projections.

## Hook kinds

| Option | Signature | Called |
|---|---|---|
| `WithBeforeInsert` | `func(ctx, *T)` | before SQL `INSERT` |
| `WithAfterInsert` | `func(ctx, *T)` | after a successful `INSERT` |
| `WithBeforeUpdate` | `func(ctx, *T)` | before SQL `UPDATE` |
| `WithAfterUpdate` | `func(ctx, *T)` | after a successful `UPDATE` (rowsAffected > 0) |
| `WithAfterSelect` | `func(ctx, []*T)` | after Scan of `GetFirst`/`GetList` |

For `GetFirst`, `afterSelect` receives a single-element slice. For `GetList` — the full slice.

## Mutating the model

Changes made in a `Before*` hook land in the SQL. This is the standard way to fill generated fields:

```go
.WithBeforeInsert(func(ctx context.Context, u *User) {
    if u.ID == uuid.Nil {
        u.ID = uuid.New()
    }
    if u.CreatedAt.IsZero() {
        u.CreatedAt = time.Now().UTC()
    }
}).
WithBeforeUpdate(func(ctx context.Context, u *User) {
    now := time.Now().UTC()
    u.UpdatedAt = &now
})
```

`After*` hooks can also mutate the model, but their changes never reach the database — they only affect the caller's copy.

## Stacking

`WithBeforeInsert` can be called multiple times — hooks run in registration order:

```go
.WithBeforeInsert(setDefaults).
WithBeforeInsert(validatePayload).
WithBeforeInsert(audit.LogAttempt)
```

## Typical uses

- **Field auto-fill:** IDs, timestamps, tenant_id.
- **Auditing:** log INSERT/UPDATE/DELETE together with the user and payload.
- **Denormalization:** recompute a derived column after UPDATE.
- **Post-processing in `AfterSelect`:** e.g. decrypting encrypted fields.

## What you can't do in a hook

- Open transactions on the same connection — you're already inside a gerpo call. Run external side effects outside `.Insert`/`.Update`.
- Return an error — the signature is `func(ctx, *T)` without a return. To abort an operation, validate upstream (before the repo call) or map the DB error via [Error transformer](error-transformer.md).
