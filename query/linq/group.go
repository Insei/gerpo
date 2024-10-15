package linq

import "github.com/insei/gerpo/sql"

func NewGroupBuilder(core *CoreBuilder) *GroupBuilder {
	return &GroupBuilder{
		core: core,
	}
}

type GroupBuilder struct {
	core *CoreBuilder
	opts []func(*sql.StringGroupBuilder)
}

func (q *GroupBuilder) Apply(b *sql.StringGroupBuilder) {
	for _, opt := range q.opts {
		opt(b)
	}
}

func (q *GroupBuilder) GroupBy(fieldsPtr ...any) {
	for _, fieldPtr := range fieldsPtr {
		col := q.core.GetColumn(fieldPtr)
		q.opts = append(q.opts, func(b *sql.StringGroupBuilder) {
			b.GroupBy(col)
		})
	}
}
