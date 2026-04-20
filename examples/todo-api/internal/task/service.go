package task

import (
	"context"

	"github.com/google/uuid"
	"github.com/insei/gerpo"
	"github.com/insei/gerpo/executor"
	"github.com/insei/gerpo/query"
)

// Service is the thin business layer above the repository. Kept small on
// purpose — most methods are one gerpo call wide. Multi-step writes go
// through gerpo.RunInTx to stay atomic across repositories, even though the
// example only has one.
type Service struct {
	repo    gerpo.Repository[Task]
	adapter executor.Adapter
}

func NewService(repo gerpo.Repository[Task], adapter executor.Adapter) *Service {
	return &Service{repo: repo, adapter: adapter}
}

// ListParams holds optional list filters. Zero values mean "no filter".
type ListParams struct {
	Page uint64
	Size uint64
	Done *bool // nil = any status
}

func (s *Service) List(ctx context.Context, p ListParams) ([]*Task, error) {
	return s.repo.GetList(ctx, func(m *Task, h query.GetListHelper[Task]) {
		if p.Done != nil {
			h.Where().Field(&m.Done).EQ(*p.Done)
		}
		h.OrderBy().Field(&m.CreatedAt).DESC()
		if p.Size == 0 {
			p.Size = 20
		}
		if p.Page == 0 {
			p.Page = 1
		}
		h.Page(p.Page).Size(p.Size)
	})
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*Task, error) {
	return s.repo.GetFirst(ctx, func(m *Task, h query.GetFirstHelper[Task]) {
		h.Where().Field(&m.ID).EQ(id)
	})
}

// Create inserts one task. The ID and CreatedAt come back from the RETURNING
// clause; the caller sees them filled in on the passed-in pointer.
func (s *Service) Create(ctx context.Context, t *Task) error {
	return s.repo.Insert(ctx, t)
}

// Update patches title / description / done for one task. Returns
// task.ErrNotFound if no row matches the id.
func (s *Service) Update(ctx context.Context, t *Task) error {
	return gerpo.RunInTx(ctx, s.adapter, func(ctx context.Context) error {
		_, err := s.repo.Update(ctx, t, func(m *Task, h query.UpdateHelper[Task]) {
			h.Where().Field(&m.ID).EQ(t.ID)
			// Only fields the API lets the caller touch — CreatedAt / ID stay put.
			h.Only(&m.Title, &m.Description, &m.Done)
		})
		return err
	})
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := s.repo.Delete(ctx, func(m *Task, h query.DeleteHelper[Task]) {
		h.Where().Field(&m.ID).EQ(id)
	})
	return err
}
