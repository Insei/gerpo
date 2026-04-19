package databasesql

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/insei/gerpo/executor/adapters/placeholder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTxFixture(t *testing.T) (*txWrap, sqlmock.Sqlmock, func()) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	mock.ExpectBegin()
	tx, err := db.Begin()
	require.NoError(t, err)
	w := &txWrap{
		tx:                            tx,
		placeholder:                   placeholder.Question,
		rollbackUnlessCommittedNeeded: true,
	}
	cleanup := func() {
		assert.NoError(t, mock.ExpectationsWereMet())
		_ = db.Close()
	}
	return w, mock, cleanup
}

// TestTxWrap_Commit_SetsCommitted verifies Commit() flips the internal flag —
// guarding against the historical "commited" typo that broke the field name
// and made RollbackUnlessCommitted() try to roll back a committed tx.
func TestTxWrap_Commit_SetsCommitted(t *testing.T) {
	w, mock, cleanup := newTxFixture(t)
	defer cleanup()

	mock.ExpectCommit()
	require.NoError(t, w.Commit())
	assert.True(t, w.committed, "Commit() must set committed=true")
}

// TestTxWrap_RollbackUnlessCommitted_AfterCommit_IsNoop ensures the safety net
// does not call the driver after a successful Commit.
func TestTxWrap_RollbackUnlessCommitted_AfterCommit_IsNoop(t *testing.T) {
	w, mock, cleanup := newTxFixture(t)
	defer cleanup()

	mock.ExpectCommit()
	require.NoError(t, w.Commit())
	// No mock.ExpectRollback() — if RollbackUnlessCommitted() called Rollback
	// the sqlmock expectation set would fail in cleanup.
	require.NoError(t, w.RollbackUnlessCommitted())
}

// TestTxWrap_RollbackUnlessCommitted_WithoutCommit_DoesRollback ensures the
// safety net rolls back when Commit was not called.
func TestTxWrap_RollbackUnlessCommitted_WithoutCommit_DoesRollback(t *testing.T) {
	w, mock, cleanup := newTxFixture(t)
	defer cleanup()

	mock.ExpectRollback()
	require.NoError(t, w.RollbackUnlessCommitted())
	assert.False(t, w.rollbackUnlessCommittedNeeded, "Rollback() must clear the safety-net flag")
}

// TestTxWrap_Rollback_ClearsSafetyNet ensures an explicit Rollback() prevents
// RollbackUnlessCommitted from trying to roll back a second time.
func TestTxWrap_Rollback_ClearsSafetyNet(t *testing.T) {
	w, mock, cleanup := newTxFixture(t)
	defer cleanup()

	mock.ExpectRollback()
	require.NoError(t, w.Rollback())
	// Second call must be a no-op — no extra ExpectRollback configured.
	require.NoError(t, w.RollbackUnlessCommitted())
}
