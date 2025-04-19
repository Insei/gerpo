package sqlstmt

import "fmt"

var (
	ErrEmptyColumnsInExecutionSet = fmt.Errorf("empty columns in execution columns set")
	ErrTableIsNoSet               = fmt.Errorf("table is not set")
)
