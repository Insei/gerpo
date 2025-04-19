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
	opts  []func(GroupApplier) error
}

type GroupApplier interface {
	ColumnsStorage() types.ColumnsStorage
	Group() sqlpart.Group
}

func (q *GroupBuilder) Apply(applier GroupApplier) error {
	for _, opt := range q.opts {
		err := opt(applier)
		if err != nil {
			return err
		}
	}
	return nil
}

func (q *GroupBuilder) GroupBy(fieldsPtr ...any) {
	for _, fieldPtr := range fieldsPtr {
		savedPtr := fieldPtr
		q.opts = append(q.opts, func(applier GroupApplier) error {
			col, err := applier.ColumnsStorage().GetByFieldPtr(q.model, savedPtr)
			if err != nil {
				return err
			}
			applier.Group().GroupBy(col)
			return nil
		})
	}
}
