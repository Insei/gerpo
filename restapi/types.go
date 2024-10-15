package restapi

import "github.com/insei/gerpo/sql"

type APIConnector interface {
	ValidateFilters(filters string) error
	AppendFilters(filters string)
	ApplyWhere(b *sql.StringWhereBuilder)
}

type APIQuery interface {
	AppendFilters(filters string) APIQuery
	AppendSorts(sorts string) APIQuery
}
