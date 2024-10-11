package query

import "github.com/insei/gerpo/sql"

type SQLApply interface {
	Apply(sqlBuilder *sql.StringBuilder)
}
