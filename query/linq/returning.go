package linq

import (
	"github.com/insei/gerpo/types"
)

// ReturningBuilder collects per-request control over the RETURNING clause.
// Default: no call → repository-level column markers (ReturnedOnInsert /
// ReturnedOnUpdate) decide what's returned. A single Returning(...) call
// overrides that — the listed fields become the returning set; an empty call
// disables RETURNING for the request altogether.
type ReturningBuilder struct {
	model     any
	applied   bool
	fieldPtrs []any
}

// ReturningApplier is the slice of stmt API the ReturningBuilder pokes at:
// resolve fields through ColumnsStorage and rewrite the stmt's returning set.
type ReturningApplier interface {
	ColumnsStorage() types.ColumnsStorage
	SetReturning(cols []types.Column)
}

func NewReturningBuilder(model any) *ReturningBuilder {
	return &ReturningBuilder{model: model}
}

// Returning narrows the RETURNING clause for this request. Calling with no
// arguments disables RETURNING; calling with explicit fields replaces the
// repository's default returning set with exactly those columns.
func (b *ReturningBuilder) Returning(fieldsPtr ...any) {
	b.applied = true
	b.fieldPtrs = append(b.fieldPtrs[:0], fieldsPtr...)
}

func (b *ReturningBuilder) Apply(applier ReturningApplier) error {
	if !b.applied {
		return nil
	}
	cols := make([]types.Column, 0, len(b.fieldPtrs))
	storage := applier.ColumnsStorage()
	for _, ptr := range b.fieldPtrs {
		col, err := storage.GetByFieldPtr(b.model, ptr)
		if err != nil {
			return err
		}
		cols = append(cols, col)
	}
	applier.SetReturning(cols)
	return nil
}
