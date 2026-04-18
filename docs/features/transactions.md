# Transactions

gerpo does not invent its own transaction layer — it works with the `Tx` returned by the adapter.

## Basic flow

```go
tx, err := adapter.BeginTx(ctx)
if err != nil {
    return err
}
defer tx.RollbackUnlessCommitted() // safety net: rolls back if we forgot to Commit

txRepo := repo.Tx(tx)

if err := txRepo.Insert(ctx, u); err != nil {
    return err // defer will roll back
}
if _, err := txRepo.Update(ctx, u, whereByID); err != nil {
    return err
}

return tx.Commit()
```

## Tx methods

| Method | Effect |
|---|---|
| `Commit() error` | Commits; all subsequent `Rollback*` calls become no-ops |
| `Rollback() error` | Explicit rollback |
| `RollbackUnlessCommitted() error` | Safe `defer`: rolls back only if Commit wasn't called |
| `ExecContext`/`QueryContext` | Raw SQL — useful when you need to bypass the repo |

## `repo.Tx(tx) Repository[T]`

This method does not open a new transaction — it **wraps** an existing one. Returns a repository bound to this `Tx`. Returns a single value, no error.

```go
orderRepo := orderRepo.Tx(tx)
itemRepo := itemRepo.Tx(tx)
// both write into the same transaction
```

## Isolation

Isolation is controlled by the driver; gerpo does not set a level. PostgreSQL defaults to Read Committed. For SERIALIZABLE/REPEATABLE READ, open the transaction directly via the adapter's `ExecContext` (`BEGIN ISOLATION LEVEL …`), or pass options via the driver's `BeginTx` (pgx accepts `pgx.TxOptions`).

## Common pitfall: multiple calls without a transaction

```go
repo.Insert(ctx, order) // on one pool connection
repo.Insert(ctx, items...) // may land on a different connection; not atomic
```

If atomicity matters — wrap in one `tx`.

## Partial rollback: savepoints

gerpo does not expose a `SAVEPOINT` API. If you need nested rollbacks, issue them via `tx.ExecContext(ctx, "SAVEPOINT sp")` / `RELEASE SAVEPOINT` / `ROLLBACK TO SAVEPOINT`.
