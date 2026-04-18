//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/insei/gerpo"
	"github.com/insei/gerpo/query"
)

// newUserRepo собирает Repository[User] с полной конфигурацией: virtual column
// post_count через LEFT JOIN, soft delete по deleted_at, persistent WHERE
// (исключение soft-deleted записей).
func newUserRepo(t *testing.T, ab adapterBundle) gerpo.Repository[User] {
	t.Helper()
	repo, err := gerpo.NewBuilder[User]().
		DB(ab.adapter).
		Table("users").
		Columns(func(m *User, c *gerpo.ColumnBuilder[User]) {
			c.Field(&m.ID).AsColumn().WithUpdateProtection()
			c.Field(&m.Name).AsColumn()
			c.Field(&m.Email).AsColumn()
			c.Field(&m.Age).AsColumn()
			c.Field(&m.CreatedAt).AsColumn().WithUpdateProtection()
			c.Field(&m.UpdatedAt).AsColumn().WithInsertProtection()
			c.Field(&m.DeletedAt).AsColumn().WithInsertProtection()
			c.Field(&m.PostCount).AsVirtual().WithSQL(func(ctx context.Context) string {
				return "COALESCE(COUNT(posts.id), 0)"
			})
		}).
		WithQuery(func(m *User, h query.PersistentHelper[User]) {
			h.LeftJoin(func(ctx context.Context) string {
				return "posts ON posts.user_id = users.id"
			})
			h.GroupBy(&m.ID, &m.Name, &m.Email, &m.Age, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt)
			h.Where().Field(&m.DeletedAt).EQ(nil)
		}).
		WithSoftDeletion(func(m *User, b *gerpo.SoftDeletionBuilder[User]) {
			b.Field(&m.DeletedAt).SetValueFn(func(ctx context.Context) any {
				return nowUTC()
			})
		}).
		Build()
	if err != nil {
		t.Fatalf("build user repo: %v", err)
	}
	return repo
}

// newPostRepo собирает Repository[Post] без persistent query и soft delete.
// Используется для тестов hooks, joins, базовый CRUD.
func newPostRepo(t *testing.T, ab adapterBundle) gerpo.Repository[Post] {
	t.Helper()
	repo, err := gerpo.NewBuilder[Post]().
		DB(ab.adapter).
		Table("posts").
		Columns(func(m *Post, c *gerpo.ColumnBuilder[Post]) {
			c.Field(&m.ID).AsColumn().WithUpdateProtection()
			c.Field(&m.UserID).AsColumn()
			c.Field(&m.Title).AsColumn()
			c.Field(&m.Content).AsColumn()
			c.Field(&m.Published).AsColumn()
			c.Field(&m.PublishedAt).AsColumn()
			c.Field(&m.CreatedAt).AsColumn().WithUpdateProtection()
		}).
		Build()
	if err != nil {
		t.Fatalf("build post repo: %v", err)
	}
	return repo
}

// newCommentRepo собирает Repository[Comment].
func newCommentRepo(t *testing.T, ab adapterBundle) gerpo.Repository[Comment] {
	t.Helper()
	repo, err := gerpo.NewBuilder[Comment]().
		DB(ab.adapter).
		Table("comments").
		Columns(func(m *Comment, c *gerpo.ColumnBuilder[Comment]) {
			c.Field(&m.ID).AsColumn().WithUpdateProtection()
			c.Field(&m.PostID).AsColumn()
			c.Field(&m.UserID).AsColumn()
			c.Field(&m.Body).AsColumn()
			c.Field(&m.CreatedAt).AsColumn().WithUpdateProtection()
		}).
		Build()
	if err != nil {
		t.Fatalf("build comment repo: %v", err)
	}
	return repo
}
