package gerpo

import (
	"context"
	"fmt"

	"github.com/insei/gerpo/executor"
)

// WithTx returns a context carrying tx. Every Repository operation invoked with
// the returned context (or any context derived from it) executes against tx
// instead of the underlying adapter — for every Repository sharing that ctx.
//
// It re-exports executor.WithTx so callers importing the gerpo package do not
// have to pull in the executor package just for transactional plumbing.
//
// Commit / Rollback / RollbackUnlessCommitted remain the caller's
// responsibility — WithTx does not lifecycle the transaction.
var WithTx = executor.WithTx

// RunInTx runs fn inside a transaction started on adapter. The transaction is
// committed when fn returns nil, rolled back otherwise. The passed-through ctx
// inside fn already carries the tx (via WithTx), so every Repository call made
// with that ctx is transactional — no manual wiring needed.
//
// If fn panics, RollbackUnlessCommitted runs; the panic is then re-raised.
func RunInTx(
	ctx context.Context,
	adapter executor.Adapter,
	fn func(ctx context.Context) error,
) (err error) {
	if adapter == nil {
		return fmt.Errorf("gerpo: RunInTx: adapter is nil")
	}
	if fn == nil {
		return fmt.Errorf("gerpo: RunInTx: fn is nil")
	}
	tx, err := adapter.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("gerpo: RunInTx: begin: %w", err)
	}
	defer func() {
		if rbErr := tx.RollbackUnlessCommitted(); rbErr != nil && err == nil {
			err = rbErr
		}
	}()

	ctx = WithTx(ctx, tx)
	if err = fn(ctx); err != nil {
		return err
	}
	return tx.Commit()
}
