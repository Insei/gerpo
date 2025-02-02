package executor

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

// By using table driven tests we can easily add more test cases in the future
func TestBeginTx(t *testing.T) {
	testCases := []struct {
		desc                 string
		opts                 []*sql.TxOptions
		beginTxErr           error
		expectedBeginTxCalls int
	}{
		{
			desc:                 "successful transaction begin without options",
			opts:                 nil,
			beginTxErr:           nil,
			expectedBeginTxCalls: 1,
		},
		{
			desc: "successful transaction begin with options",
			opts: []*sql.TxOptions{{
				Isolation: sql.LevelReadCommitted,
				ReadOnly:  true,
			}},
			beginTxErr:           nil,
			expectedBeginTxCalls: 1,
		},
		{
			desc:                 "failed transaction begin with DB error",
			opts:                 nil,
			beginTxErr:           errors.New("DB error"),
			expectedBeginTxCalls: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			// Use sqlmock to create a mock DB
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			// Here we setup our mock DB to act as we want it to
			if tc.beginTxErr != nil {
				mock.ExpectBegin().WillReturnError(tc.beginTxErr)
			} else {
				mock.ExpectBegin()
			}

			// Run the function and check that the error is as expected
			_, err = BeginTx(context.Background(), db, tc.opts...)

			if tc.beginTxErr != nil {
				require.Error(t, err)
				require.Equal(t, tc.beginTxErr, err)
			} else {
				require.NoError(t, err)
			}

			// Check that all expected calls were made
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// Test for Commit method
func TestCommit(t *testing.T) {
	testCases := []struct {
		description string
		commitErr   error
		expectedErr error
	}{
		{
			description: "successful commit",
			commitErr:   nil,
			expectedErr: nil,
		},
		{
			description: "failed commit",
			commitErr:   errors.New("commit error"),
			expectedErr: errors.New("commit error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			mock.ExpectBegin()
			if tc.commitErr != nil {
				mock.ExpectCommit().WillReturnError(tc.commitErr)
			} else {
				mock.ExpectCommit()
			}

			sqlTx, err := db.Begin()
			require.NoError(t, err)

			tx := &Tx{tx: sqlTx}

			err = tx.Commit()
			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.True(t, tx.commited)
			}
		})
	}
}

// Test for Rollback method
func TestRollback(t *testing.T) {
	testCases := []struct {
		description  string
		rollbackErr  error
		expectedErr  error
		initialFlag  bool
		expectedFlag bool
	}{
		{
			description:  "successful rollback",
			rollbackErr:  nil,
			expectedErr:  nil,
			initialFlag:  true,
			expectedFlag: false,
		},
		{
			description:  "failed rollback",
			rollbackErr:  errors.New("rollback error"),
			expectedErr:  errors.New("rollback error"),
			initialFlag:  true,
			expectedFlag: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			mock.ExpectBegin()
			if tc.rollbackErr != nil {
				mock.ExpectRollback().WillReturnError(tc.rollbackErr)
			} else {
				mock.ExpectRollback()
			}

			sqlTx, err := db.Begin()
			require.NoError(t, err)

			tx := &Tx{
				tx:                            sqlTx,
				rollbackUnlessCommittedNeeded: tc.initialFlag,
			}
			err = tx.Rollback()
			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.expectedFlag, tx.rollbackUnlessCommittedNeeded)
		})
	}
}

// Test for RollbackUnlessCommitted method
func TestRollbackUnlessCommitted(t *testing.T) {
	testCases := []struct {
		description    string
		committed      bool
		rollbackNeeded bool
		rollbackErr    error
		shouldPanic    bool
	}{
		{
			description:    "no rollback if committed",
			committed:      true,
			rollbackNeeded: true,
			rollbackErr:    nil,
			shouldPanic:    false,
		},
		{
			description:    "no rollback if not needed",
			committed:      false,
			rollbackNeeded: false,
			rollbackErr:    nil,
			shouldPanic:    false,
		},
		{
			description:    "rollback without error",
			committed:      false,
			rollbackNeeded: true,
			rollbackErr:    nil,
			shouldPanic:    false,
		},
		{
			description:    "rollback with error",
			committed:      false,
			rollbackNeeded: true,
			rollbackErr:    errors.New("rollback error"),
			shouldPanic:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			mock.ExpectBegin()
			if tc.rollbackErr != nil {
				mock.ExpectRollback().WillReturnError(tc.rollbackErr)
			} else {
				mock.ExpectRollback()
			}

			sqlTx, err := db.Begin()
			require.NoError(t, err)

			tx := &Tx{
				tx:                            sqlTx,
				commited:                      tc.committed,
				rollbackUnlessCommittedNeeded: tc.rollbackNeeded,
			}

			if tc.shouldPanic {
				require.Panics(t, func() {
					tx.RollbackUnlessCommitted()
				})
			} else {
				tx.RollbackUnlessCommitted()
			}
		})
	}
}
