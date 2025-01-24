package linq

import "github.com/insei/gerpo/sqlstmt/sqlpart"

type PaginationApplier interface {
	LimitOffset() sqlpart.LimitOffset
}

type PaginationBuilder struct {
	limit uint64
	page  uint64
}

func NewPaginationBuilder() *PaginationBuilder {
	return &PaginationBuilder{
		limit: 10,
		page:  0,
	}
}

func (q *PaginationBuilder) Page(page uint64) {
	if page > 0 {
		page--
	}
	q.page = page
}

func (q *PaginationBuilder) Size(size uint64) {
	q.limit = size
}

func (q *PaginationBuilder) Apply(applier PaginationApplier) {
	applier.LimitOffset().SetLimit(q.limit)
	applier.LimitOffset().SetOffset(q.page * q.limit)
}
