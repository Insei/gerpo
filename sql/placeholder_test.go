package sql

import (
	"testing"

	"github.com/insei/fmap/v3"
	"github.com/stretchr/testify/assert"
)

func TestDetermineByConnectorTypeName(t *testing.T) {
	testCases := []struct {
		name     string
		typeName string
	}{
		{
			name:     "Contains stdlib",
			typeName: "pgx.Conn",
		},
		{
			name:     "Does not contain stdlib",
			typeName: "mysql.Conn",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			determineByConnectorTypeName(tc.typeName)
		})
	}
}

func TestDetermineByDriverName(t *testing.T) {
	testCases := []struct {
		name       string
		driverName string
	}{
		{
			name:       "Contains pq.Driver",
			driverName: "pq.Driver",
		},
		{
			name:       "Contains stdlib.Driver",
			driverName: "stdlib.Driver",
		},
		{
			name:       "Does not contain pq.Driver or stdlib.Driver",
			driverName: "mysql.Driver",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			determineByDriverName(tc.driverName)
		})
	}
}

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

func TestUnwrapConnector(t *testing.T) {
	type mockConnector struct {
		Connector any
	}

	testCases := []struct {
		name      string
		connector any
	}{
		{
			name:      "Connector fieldType found",
			connector: mockConnector{Connector: "test connector value"},
		},
		{
			name:      "Connector fieldType not found",
			connector: struct{}{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockStorage, _ := fmap.GetFrom(tc.connector)
			unwrapConnector(tc.connector, mockStorage)
		})
	}
}
