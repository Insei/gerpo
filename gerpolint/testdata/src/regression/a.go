package regression

import (
	"time"

	"github.com/insei/gerpo/types"
)

// Task mirrors examples/todo-api shape: uuid-ish ID, timestamps, typed bools.
type Task struct {
	ID        string
	Title     string
	Done      bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type helper struct{}

func (helper) Where() types.WhereTarget { return whereTarget{} }

type whereTarget struct{}

func (whereTarget) Field(_ any) types.WhereOperation            { return nil }
func (whereTarget) Group(_ func(types.WhereTarget)) types.ANDOR { return &andor{} }

type andor struct{}

func (*andor) AND() types.WhereTarget { return whereTarget{} }
func (*andor) OR() types.WhereTarget  { return whereTarget{} }

// idiomaticGerpoCode should produce zero diagnostics. Snapshot of code shapes
// lifted from examples/todo-api/internal/task/service.go — if a future change
// to the analyzer breaks this, we have a regression signal immediately.
func idiomaticGerpoCode(id string, donePtr *bool) {
	h := helper{}
	m := &Task{}

	h.Where().Field(&m.ID).EQ(id)

	if donePtr != nil {
		h.Where().Field(&m.Done).EQ(*donePtr)
	}

	h.Where().Field(&m.Title).Contains("urgent")
	h.Where().Field(&m.CreatedAt).GTE(time.Now())
}

func groupedFilters(since time.Time) {
	h := helper{}
	m := &Task{}

	h.Where().Group(func(t types.WhereTarget) {
		t.Field(&m.Done).EQ(false)
		t.Field(&m.UpdatedAt).LT(since)
	})
}
