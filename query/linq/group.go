package linq

import (
	"github.com/insei/gerpo/sqlstmt/sqlpart"
	"github.com/insei/gerpo/types"
)

func NewGroupBuilder(baseModel any) *GroupBuilder {
	return &GroupBuilder{
		model: baseModel,
	}
}

type GroupBuilder struct {
	model any
	opts  []func(GroupApplier)
}

type GroupApplier interface {
	ColumnsStorage() *types.ColumnsStorage
	Group() sqlpart.Group
}

func (q *GroupBuilder) Apply(applier GroupApplier) {
	for _, opt := range q.opts {
		opt(applier)
	}
}

func (q *GroupBuilder) GroupBy(fieldsPtr ...any) {
	for _, fieldPtr := range fieldsPtr {
		q.opts = append(q.opts, func(applier GroupApplier) {
			col, err := applier.ColumnsStorage().GetByFieldPtr(q.model, fieldPtr)
			if err != nil {
				panic(err)
			}
			applier.Group().GroupBy(col)
		})
	}
}
