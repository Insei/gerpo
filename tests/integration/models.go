//go:build integration

package integration

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Name      string
	Email     *string
	Age       int
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
	// Virtual column: COUNT(posts.id) за пользователя (считается через LEFT JOIN).
	PostCount int
}

type Post struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Title       string
	Content     string
	Published   bool
	PublishedAt *time.Time
	CreatedAt   time.Time
}

type Comment struct {
	ID        uuid.UUID
	PostID    uuid.UUID
	UserID    uuid.UUID
	Body      string
	CreatedAt time.Time
}
