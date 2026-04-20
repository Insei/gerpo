package internal

import (
	"context"
	"errors"
	"testing"

	"github.com/insei/gerpo/executor/adapters/placeholder"
	extypes "github.com/insei/gerpo/executor/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeDriver records every call and lets tests choose what to return.
type fakeDriver struct {
	execCalls  []callRecord
	queryCalls []callRecord
	beginErr   error
	tx         *fakeTx
	execResult extypes.Result
	queryRows  extypes.Rows
	execErr    error
	queryErr   error
}

type callRecord struct {
	sql  string
	args []any
}

func (b *fakeDriver) Exec(_ context.Context, sql string, args ...any) (extypes.Result, error) {
	b.execCalls = append(b.execCalls, callRecord{sql: sql, args: args})
	return b.execResult, b.execErr
}

func (b *fakeDriver) Query(_ context.Context, sql string, args ...any) (extypes.Rows, error) {
	b.queryCalls = append(b.queryCalls, callRecord{sql: sql, args: args})
	return b.queryRows, b.queryErr
}

func (b *fakeDriver) BeginTx(_ context.Context) (TxDriver, error) {
	if b.beginErr != nil {
		return nil, b.beginErr
	}
	if b.tx == nil {
		b.tx = &fakeTx{}
	}
	return b.tx, nil
}

// fakeTx records lifecycle calls and lets tests inject errors.
type fakeTx struct {
	execCalls   []callRecord
	queryCalls  []callRecord
	commits     int
	rollbacks   int
	commitErr   error
	rollbackErr error
}

func (t *fakeTx) Exec(_ context.Context, sql string, args ...any) (extypes.Result, error) {
	t.execCalls = append(t.execCalls, callRecord{sql: sql, args: args})
	return nil, nil
}
func (t *fakeTx) Query(_ context.Context, sql string, args ...any) (extypes.Rows, error) {
	t.queryCalls = append(t.queryCalls, callRecord{sql: sql, args: args})
	return nil, nil
}
func (t *fakeTx) Commit() error {
	if t.commitErr != nil {
		return t.commitErr
	}
	t.commits++
	return nil
}
func (t *fakeTx) Rollback() error {
	if t.rollbackErr != nil {
		return t.rollbackErr
	}
	t.rollbacks++
	return nil
}

// TestAdapter_RewritesQuestionToDollar — placeholders are converted before
// reaching the driver. Drives the real placeholder.Dollar transformation.
func TestAdapter_RewritesQuestionToDollar(t *testing.T) {
	b := &fakeDriver{}
	a := New(b, placeholder.Dollar)

	_, err := a.ExecContext(context.Background(), "INSERT INTO t(a, b) VALUES (?, ?)", "x", 1)
	require.NoError(t, err)
	require.Len(t, b.execCalls, 1)
	assert.Equal(t, "INSERT INTO t(a, b) VALUES ($1, $2)", b.execCalls[0].sql)
	assert.Equal(t, []any{"x", 1}, b.execCalls[0].args)

	_, err = a.QueryContext(context.Background(), "SELECT 1 FROM t WHERE a = ? AND b = ?", "y", 2)
	require.NoError(t, err)
	require.Len(t, b.queryCalls, 1)
	assert.Equal(t, "SELECT 1 FROM t WHERE a = $1 AND b = $2", b.queryCalls[0].sql)
}

// TestAdapter_QuestionPlaceholder_NoOp — Question format leaves SQL alone.
func TestAdapter_QuestionPlaceholder_NoOp(t *testing.T) {
	b := &fakeDriver{}
	a := New(b, placeholder.Question)

	_, err := a.ExecContext(context.Background(), "INSERT INTO t(a) VALUES (?)", "x")
	require.NoError(t, err)
	assert.Equal(t, "INSERT INTO t(a) VALUES (?)", b.execCalls[0].sql)
}

// TestTransaction_Commit_FlipsCommittedFlag — Commit sets the internal flag so
// subsequent RollbackUnlessCommitted is a no-op.
func TestTransaction_Commit_FlipsCommittedFlag(t *testing.T) {
	b := &fakeDriver{}
	a := New(b, placeholder.Question)

	tx, err := a.BeginTx(context.Background())
	require.NoError(t, err)

	require.NoError(t, tx.Commit())
	assert.Equal(t, 1, b.tx.commits)

	require.NoError(t, tx.RollbackUnlessCommitted(),
		"after Commit RollbackUnlessCommitted must be a no-op")
	assert.Equal(t, 0, b.tx.rollbacks, "no rollback should reach the driver after a successful Commit")
}

// TestTransaction_RollbackUnlessCommitted_WithoutCommit_RollsBack — happy path
// for the safety net.
func TestTransaction_RollbackUnlessCommitted_WithoutCommit_RollsBack(t *testing.T) {
	b := &fakeDriver{}
	a := New(b, placeholder.Question)

	tx, err := a.BeginTx(context.Background())
	require.NoError(t, err)

	require.NoError(t, tx.RollbackUnlessCommitted())
	assert.Equal(t, 1, b.tx.rollbacks)

	// Second call must not roll back again — the safety net flag is cleared.
	require.NoError(t, tx.RollbackUnlessCommitted())
	assert.Equal(t, 1, b.tx.rollbacks)
}

// TestTransaction_ExplicitRollback_ClearsSafetyNet — Rollback by itself also
// blocks the deferred RollbackUnlessCommitted.
func TestTransaction_ExplicitRollback_ClearsSafetyNet(t *testing.T) {
	b := &fakeDriver{}
	a := New(b, placeholder.Question)

	tx, err := a.BeginTx(context.Background())
	require.NoError(t, err)

	require.NoError(t, tx.Rollback())
	require.NoError(t, tx.RollbackUnlessCommitted())
	assert.Equal(t, 1, b.tx.rollbacks, "second rollback through the safety net must not reach the driver")
}

// TestTransaction_CommitError_DoesNotMarkCommitted — if Commit fails the flag
// stays false so RollbackUnlessCommitted will still try to roll back.
func TestTransaction_CommitError_DoesNotMarkCommitted(t *testing.T) {
	commitFail := errors.New("commit failed")
	b := &fakeDriver{tx: &fakeTx{commitErr: commitFail}}
	a := New(b, placeholder.Question)

	tx, err := a.BeginTx(context.Background())
	require.NoError(t, err)

	assert.ErrorIs(t, tx.Commit(), commitFail)
	require.NoError(t, tx.RollbackUnlessCommitted())
	assert.Equal(t, 1, b.tx.rollbacks, "failed Commit must leave the safety net armed")
}

// TestTransaction_ExecAndQuery_RewritePlaceholders — transactional
// Exec/Query also pass through the placeholder rewriter.
func TestTransaction_ExecAndQuery_RewritePlaceholders(t *testing.T) {
	b := &fakeDriver{}
	a := New(b, placeholder.Dollar)

	tx, err := a.BeginTx(context.Background())
	require.NoError(t, err)

	_, err = tx.ExecContext(context.Background(), "UPDATE t SET a = ? WHERE id = ?", "x", 1)
	require.NoError(t, err)
	assert.Equal(t, "UPDATE t SET a = $1 WHERE id = $2", b.tx.execCalls[0].sql)

	_, err = tx.QueryContext(context.Background(), "SELECT * FROM t WHERE id = ?", 1)
	require.NoError(t, err)
	assert.Equal(t, "SELECT * FROM t WHERE id = $1", b.tx.queryCalls[0].sql)
}

// TestAdapter_BeginTxError_Propagates — driver BeginTx errors reach the
// caller as-is.
func TestAdapter_BeginTxError_Propagates(t *testing.T) {
	beginFail := errors.New("begin failed")
	b := &fakeDriver{beginErr: beginFail}
	a := New(b, placeholder.Question)

	tx, err := a.BeginTx(context.Background())
	assert.Nil(t, tx)
	assert.ErrorIs(t, err, beginFail)
}
