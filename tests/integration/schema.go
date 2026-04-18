//go:build integration

package integration

import (
	_ "embed"
)

//go:embed schema.sql
var schemaSQL string
