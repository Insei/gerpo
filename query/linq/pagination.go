package linq

type Limiter interface {
	Limit(uint64)
	Offset(uint64)
}

type PaginationBuilder struct {
	opts  []func(b Limiter)
	limit uint64
	page  uint64
}

func NewPaginationBuilder() *PaginationBuilder {
	return &PaginationBuilder{
		limit: 1,
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

func (q *PaginationBuilder) Apply(paginator Limiter) {
	q.opts = append(q.opts, func(b Limiter) {
		b.Limit(q.limit)
		b.Offset(q.page * q.limit)
	})
	for _, opt := range q.opts {
		opt(paginator)
	}
}
