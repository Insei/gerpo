package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostgres(t *testing.T) {
	testCases := []struct {
		name        string
		sql         string
		expectedSQL string
	}{
		{
			name:        "Empty input",
			sql:         "",
			expectedSQL: "",
		},
		{
			name:        "Single placeholder",
			sql:         "SELECT * FROM table WHERE id = ?",
			expectedSQL: "SELECT * FROM table WHERE id = $1",
		},
		{
			name:        "Multiple placeholders",
			sql:         "SELECT * FROM table WHERE id = ? AND name = ?",
			expectedSQL: "SELECT * FROM table WHERE id = $1 AND name = $2",
		},
		{
			name:        "No placeholder",
			sql:         "SELECT * FROM table",
			expectedSQL: "SELECT * FROM table",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := postgres(tc.sql, 1)
			assert.Equal(t, tc.expectedSQL, result)
		})
	}
}
