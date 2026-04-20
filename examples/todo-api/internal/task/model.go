package task

import (
	"time"

	"github.com/google/uuid"
)

// Task is the domain model bound to the `tasks` table. Field bindings happen
// in repo.go via pointers — no struct tags required.
type Task struct {
	ID          uuid.UUID
	Title       string
	Description string
	Done        bool
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}
