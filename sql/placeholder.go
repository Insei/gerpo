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

func determineByConnectorTypeName(typeName string) Placeholder {
	if strings.Contains(typeName, "stdlib") {
		return PostgresPlaceholder
	}
	return NoopPlaceholder
}

func determineByDriverName(driverName string) Placeholder {
	if strings.Contains(driverName, "pq.Driver") || strings.Contains(driverName, "stdlib.Driver") {
		return PostgresPlaceholder
	}
	return NoopPlaceholder
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
	connectorTypeStr := fmt.Sprintf("%T", connector)
	connectorFields, err := fmap.GetFrom(connector)
	if err != nil {
		return determineByConnectorTypeName(connectorTypeStr)
	}
	driverField, ok := connectorFields.Find("driver")
	if !ok {
		return NoopPlaceholder
	}
	driver := driverField.Get(connector)
	wrappedDriverFields, err := fmap.GetFrom(driver)
	if err == nil {
		dwrappedDriverField, ok := wrappedDriverFields.Find("driver")
		if ok {
			driver = dwrappedDriverField.Get(db)
		}
	}
	return determineByDriverName(fmt.Sprintf("%T", driver))
}
