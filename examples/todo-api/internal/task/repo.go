package task

import (
	"errors"

	"github.com/insei/gerpo"
	"github.com/insei/gerpo/executor"
	cachectx "github.com/insei/gerpo/executor/cache/ctx"
)

// ErrNotFound is the domain-level "missing row" error. Repository-level
// gerpo.ErrNotFound is translated to this via WithErrorTransformer so HTTP
// handlers do not have to know about gerpo at all.
var ErrNotFound = errors.New("task not found")

// NewRepository builds the gerpo Repository[Task] for a given adapter.
//
// Every repo operation is wrapped in a request-scope cache (handed to every
// request via CacheMiddleware in cmd/server) and domain errors are mapped
// before leaving the layer.
func NewRepository(adapter executor.Adapter) (gerpo.Repository[Task], error) {
	return gerpo.New[Task]().
		Adapter(adapter, executor.WithCacheStorage(cachectx.New())).
		Table("tasks").
		Columns(func(m *Task, c *gerpo.ColumnBuilder[Task]) {
			// Server-generated identifiers / timestamps — read back via RETURNING.
			c.Field(&m.ID).ReadOnly().ReturnedOnInsert()
			c.Field(&m.CreatedAt).ReadOnly().ReturnedOnInsert()
			// Trigger-maintained column — written on every UPDATE by the server.
			c.Field(&m.UpdatedAt).OmitOnInsert().ReturnedOnUpdate()
			c.Field(&m.Title).AsVirtual()
			c.Field(&m.Description)
			c.Field(&m.Done)
		}).
		WithErrorTransformer(func(err error) error {
			if errors.Is(err, gerpo.ErrNotFound) {
				return ErrNotFound
			}
			return err
		}).
		Build()
}
