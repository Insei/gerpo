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
	model     any
	fieldPtrs []any
}

type GroupApplier interface {
	ColumnsStorage() types.ColumnsStorage
	Group() sqlpart.Group
}

func (q *GroupBuilder) Apply(applier GroupApplier) error {
	if len(q.fieldPtrs) == 0 {
		return nil
	}
	storage := applier.ColumnsStorage()
	group := applier.Group()
	for _, fieldPtr := range q.fieldPtrs {
		col, err := storage.GetByFieldPtr(q.model, fieldPtr)
		if err != nil {
			return err
		}
		group.GroupBy(col)
	}
	return nil
}

func (q *GroupBuilder) GroupBy(fieldsPtr ...any) {
	q.fieldPtrs = append(q.fieldPtrs, fieldsPtr...)
}
