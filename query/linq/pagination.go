package linq

type Limiter interface {
	Limit(uint64)
	Offset(uint64)
}

type PaginationBuilder struct {
	opts []func(b Limiter)
}

func NewPaginationBuilder() *PaginationBuilder {
	return &PaginationBuilder{}
}

func (q *PaginationBuilder) Page(page uint64) {
	q.opts = append(q.opts, func(b Limiter) {
		if page > 0 {
			page--
		}
		b.Offset(page)
	})
}

func (q *PaginationBuilder) Size(size uint64) {
	q.opts = append(q.opts, func(b Limiter) {
		if size == 0 {
			size = 10
		}
		b.Limit(size)
	})
}

func (q *PaginationBuilder) Apply(paginator Limiter) {
	for _, opt := range q.opts {
		opt(paginator)
	}
}
