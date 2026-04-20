package executor

import (
	"context"

	extypes "github.com/insei/gerpo/executor/types"
)

// txKey is the private key gerpo uses to stash a Tx into a context.Context.
// Keeping it unexported prevents foreign packages from constructing values that
// would shadow the real transaction.
type txKey struct{}

// WithTx returns a derived context carrying tx. Every Repository operation
// invoked with the returned context (or any context derived from it) runs
// against tx instead of the raw adapter — no matter which Repository receives
// the call. Commit / Rollback / RollbackUnlessCommitted remain the caller's
// responsibility and are invoked on tx directly.
//
// Behavior is first-tx-wins: calling WithTx a second time on an already-
// tx-bearing context replaces the stored transaction. Passing a nil tx returns
// the ctx unchanged.
func WithTx(ctx context.Context, tx extypes.Tx) context.Context {
	if tx == nil {
		return ctx
	}
	return context.WithValue(ctx, txKey{}, tx)
}

// txFromContext retrieves a Tx previously installed by WithTx. The boolean
// signals whether the context actually carries one — a plain nil check on the
// Tx is not enough because a typed-nil interface would read as non-nil.
func txFromContext(ctx context.Context) (extypes.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(extypes.Tx)
	return tx, ok && tx != nil
}
