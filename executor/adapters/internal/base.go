// Package internal hosts the placeholder-rewriting plumbing shared by every
// bundled gerpo DB adapter (pgx v5, pgx v4, database/sql). Each driver only
// needs to provide a tiny Backend implementation; the plumbing — placeholder
// rewrite, transaction state machine, RollbackUnlessCommitted semantics —
// lives here so all adapters stay consistent.
//
// The package is internal: it is not part of the public API surface.
package internal

import (
	"context"
	"fmt"

	"github.com/insei/gerpo/executor/adapters/placeholder"
	extypes "github.com/insei/gerpo/executor/types"
)

// Backend describes the driver-specific behavior the generic Adapter wraps.
// Implementations are expected to convert their native Result/Rows types into
// the executor.types interfaces themselves — Backend works in already-rewritten
// SQL, the generic Adapter handles placeholder translation upstream.
type Backend interface {
	Exec(ctx context.Context, sql string, args ...any) (extypes.Result, error)
	Query(ctx context.Context, sql string, args ...any) (extypes.Rows, error)
	BeginTx(ctx context.Context) (TxBackend, error)
}

// TxBackend mirrors Backend minus BeginTx, plus Commit / Rollback. The wrapping
// transaction owns the committed / rollbackUnlessCommittedNeeded flags so
// drivers do not need to reimplement them.
type TxBackend interface {
	Exec(ctx context.Context, sql string, args ...any) (extypes.Result, error)
	Query(ctx context.Context, sql string, args ...any) (extypes.Rows, error)
	Commit() error
	Rollback() error
}

// Adapter is the executor.types.DBAdapter implementation shared by every
// bundled driver. It rewrites placeholders before each driver call and wraps
// transactions in a state machine that makes RollbackUnlessCommitted safe to
// use as a defer.
type Adapter struct {
	backend     Backend
	placeholder placeholder.PlaceholderFormat
}

// New constructs an Adapter that runs every SQL statement through the given
// placeholder format before handing it over to the backend.
func New(backend Backend, p placeholder.PlaceholderFormat) extypes.DBAdapter {
	return &Adapter{backend: backend, placeholder: p}
}

func (a *Adapter) ExecContext(ctx context.Context, sql string, args ...any) (extypes.Result, error) {
	rewritten, err := a.placeholder.ReplacePlaceholders(sql)
	if err != nil {
		return nil, fmt.Errorf("failed to replace placeholders: %w", err)
	}
	return a.backend.Exec(ctx, rewritten, args...)
}

func (a *Adapter) QueryContext(ctx context.Context, sql string, args ...any) (extypes.Rows, error) {
	rewritten, err := a.placeholder.ReplacePlaceholders(sql)
	if err != nil {
		return nil, fmt.Errorf("failed to replace placeholders: %w", err)
	}
	return a.backend.Query(ctx, rewritten, args...)
}

func (a *Adapter) BeginTx(ctx context.Context) (extypes.Tx, error) {
	inner, err := a.backend.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	return &transaction{
		inner:                         inner,
		placeholder:                   a.placeholder,
		rollbackUnlessCommittedNeeded: true,
	}, nil
}

type transaction struct {
	inner                         TxBackend
	placeholder                   placeholder.PlaceholderFormat
	committed                     bool
	rollbackUnlessCommittedNeeded bool
}

func (t *transaction) ExecContext(ctx context.Context, sql string, args ...any) (extypes.Result, error) {
	rewritten, err := t.placeholder.ReplacePlaceholders(sql)
	if err != nil {
		return nil, fmt.Errorf("failed to replace placeholders: %w", err)
	}
	return t.inner.Exec(ctx, rewritten, args...)
}

func (t *transaction) QueryContext(ctx context.Context, sql string, args ...any) (extypes.Rows, error) {
	rewritten, err := t.placeholder.ReplacePlaceholders(sql)
	if err != nil {
		return nil, fmt.Errorf("failed to replace placeholders: %w", err)
	}
	return t.inner.Query(ctx, rewritten, args...)
}

func (t *transaction) Commit() error {
	if err := t.inner.Commit(); err != nil {
		return err
	}
	t.committed = true
	return nil
}

func (t *transaction) Rollback() error {
	t.rollbackUnlessCommittedNeeded = false
	return t.inner.Rollback()
}

func (t *transaction) RollbackUnlessCommitted() error {
	if !t.committed && t.rollbackUnlessCommittedNeeded {
		return t.Rollback()
	}
	return nil
}
