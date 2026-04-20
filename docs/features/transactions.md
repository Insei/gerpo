# Transactions

gerpo does not invent its own transaction layer — it works with the `Tx` returned by the adapter. Tx is propagated to repositories through `context.Context`, so multiple repositories can share the same transaction just by sharing the same ctx.

## Basic flow — manual

```go
tx, err := adapter.BeginTx(ctx)
if err != nil {
    return err
}
defer tx.RollbackUnlessCommitted() // safety net: rolls back if Commit wasn't called

txCtx := gerpo.WithTx(ctx, tx)     // inject tx into ctx

if err := userRepo.Insert(txCtx, u); err != nil {
    return err                     // defer will roll back
}
if _, err := userRepo.Update(txCtx, u, whereByID); err != nil {
    return err
}

return tx.Commit()
```

Any `Repository` method invoked with `txCtx` — or a context derived from it — runs against the transaction, regardless of which repository is called. A single `WithTx` covers `userRepo`, `orderRepo`, `itemRepo` at once.

## Higher-level form — `gerpo.RunInTx`

For the common "do some work in a transaction, commit on success, rollback on error" shape:

```go
err := gerpo.RunInTx(ctx, adapter, func(ctx context.Context) error {
    if err := orderRepo.Insert(ctx, order); err != nil {
        return err
    }
    for _, item := range items {
        if err := itemRepo.Insert(ctx, &item); err != nil {
            return err
        }
    }
    return nil
})
```

`RunInTx` begins the transaction, injects it into the ctx it passes to `fn`, and commits / rolls back based on the error returned from `fn`. A panic inside `fn` is propagated after `RollbackUnlessCommitted` runs.

## Tx methods

| Method | Effect |
|---|---|
| `Commit() error` | Commits; subsequent `Rollback*` calls become no-ops |
| `Rollback() error` | Explicit rollback |
| `RollbackUnlessCommitted() error` | Safe `defer`: rolls back only if Commit wasn't called |
| `ExecContext`/`QueryContext` | Raw SQL — useful when you need to bypass the repo |

## Isolation

Isolation is controlled by the driver; gerpo does not set a level. PostgreSQL defaults to Read Committed. For SERIALIZABLE/REPEATABLE READ, open the transaction directly via the adapter's `ExecContext` (`BEGIN ISOLATION LEVEL …`), or pass options via the driver's `BeginTx` (pgx accepts `pgx.TxOptions`).

## Cascading related rows

Combining a transaction with an `AfterInsert`/`AfterUpdate` hook is how gerpo lets you express user-land one-to-many relations — the hook inserts children through their own repository, both the parent and the children land in the same tx, any failure rolls everything back. The pattern lives in the hooks page: [Hooks → Cascading related rows](hooks.md#cascading-related-rows-user-land-one-to-many).

## Common pitfall: multiple calls without a transaction

```go
repo.Insert(ctx, order)       // on one pool connection
repo.Insert(ctx, items...)    // may land on a different connection; not atomic
```

If atomicity matters — wrap in one `tx` and inject through `gerpo.WithTx`.

## One adapter per context

`gerpo.WithTx` stores the transaction in ctx without remembering which adapter produced it. If you accidentally pass a ctx carrying a tx from adapter A to a repository built on adapter B, the tx will be used anyway — and the driver on the other end will reject the alien connection. In practice every app has one adapter per process, so this is a theoretical concern; be cautious if you mix adapters in the same process.

## Partial rollback: savepoints

gerpo does not expose a dedicated `SAVEPOINT` API today. If you need nested rollbacks, issue them via raw SQL on the tx:

```go
_, _ = tx.ExecContext(ctx, "SAVEPOINT sp_one")
// ... some work that may fail ...
_, _ = tx.ExecContext(ctx, "ROLLBACK TO SAVEPOINT sp_one")
// or on success:
_, _ = tx.ExecContext(ctx, "RELEASE SAVEPOINT sp_one")
```

!!! info "Roadmap"
    A first-class savepoint API is listed in [`TODO.md`](https://github.com/insei/gerpo/blob/main/TODO.md). If you need it — please [open an issue](https://github.com/insei/gerpo/issues) describing your use case so we can shape the API around real requirements.
