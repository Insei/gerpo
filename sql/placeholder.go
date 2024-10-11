package sql

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/insei/fmap/v3"
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

func DeterminePlaceHolder(db *sql.DB) Placeholder {
	dbFields, err := fmap.GetFrom(db)
	if err != nil {
		return NoopPlaceholder
	}
	connectorField, ok := dbFields.Find("connector")
	if !ok {
		return NoopPlaceholder
	}
	connector := connectorField.Get(db)
	if strings.Contains(fmt.Sprintf("%T", connector), "stdlib") {
		return PostgresPlaceholder
	}
	return NoopPlaceholder
}
