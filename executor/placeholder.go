package executor

import (
	"database/sql"
	"fmt"
	"strings"
)

type Placeholder func(sql string) string

var NoopPlaceholder = Placeholder(func(sql string) string {
	return sql
})

var PostgresPlaceholder = Placeholder(func(sql string) string {
	return postgres(sql, 1)
})

func postgres(sql string, i int) string {
	ind := strings.Index(sql, "?")
	if ind == -1 {
		return sql
	}
	sql = sql[:ind] + fmt.Sprintf("$%d", i) + sql[ind+1:]
	i++
	return postgres(sql, i)
}

func determinePlaceHolder(_ *sql.DB) Placeholder {
	return PostgresPlaceholder
}
