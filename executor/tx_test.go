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
