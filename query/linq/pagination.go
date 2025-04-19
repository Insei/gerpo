package linq

import (
	"fmt"

	"github.com/insei/gerpo/sqlstmt/sqlpart"
)

type PaginationApplier interface {
	LimitOffset() sqlpart.LimitOffset
}

type PaginationBuilder struct {
	size uint64
	page uint64
}

func NewPaginationBuilder() *PaginationBuilder {
	return &PaginationBuilder{
		size: 0,
		page: 0,
	}
}

func (q *PaginationBuilder) Page(page uint64) {
	q.page = page
}

func (q *PaginationBuilder) Size(size uint64) {
	q.size = size
}

func (q *PaginationBuilder) Apply(applier PaginationApplier) error {
	applier.LimitOffset().SetLimit(q.size)
	if q.page != 0 && q.size == 0 {
		return fmt.Errorf("incorrect pagination: size is required then page is set")
	}
	applier.LimitOffset().SetOffset((q.page - 1) * q.size)
	return nil
}
