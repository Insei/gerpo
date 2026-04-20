package sqlpart

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinBuilder_Bound(t *testing.T) {
	b := NewJoinBuilder(context.Background())
	b.JOINOn("LEFT JOIN posts ON posts.user_id = users.id AND posts.tenant_id = ?", "tenant-1")
	b.JOINOn("INNER JOIN tags ON tags.post_id = posts.id AND tags.kind IN (?, ?)", "blog", "draft")

	assert.Equal(t,
		" LEFT JOIN posts ON posts.user_id = users.id AND posts.tenant_id = ?"+
			" INNER JOIN tags ON tags.post_id = posts.id AND tags.kind IN (?, ?)",
		b.SQL())
	assert.Equal(t, []any{"tenant-1", "blog", "draft"}, b.Values())
}

func TestJoinBuilder_NoArgs(t *testing.T) {
	b := NewJoinBuilder(context.Background())
	b.JOINOn("INNER JOIN sessions ON sessions.user_id = users.id")
	b.JOINOn("LEFT JOIN posts ON posts.user_id = users.id AND posts.tenant_id = ?", "tenant-7")

	assert.Equal(t,
		" INNER JOIN sessions ON sessions.user_id = users.id"+
			" LEFT JOIN posts ON posts.user_id = users.id AND posts.tenant_id = ?",
		b.SQL())
	assert.Equal(t, []any{"tenant-7"}, b.Values())
}

func TestJoinBuilder_EmptySQL_Skipped(t *testing.T) {
	b := NewJoinBuilder(context.Background())
	b.JOINOn("")
	b.JOINOn("INNER JOIN tags ON tags.id = posts.tag_id")

	assert.Equal(t, " INNER JOIN tags ON tags.id = posts.tag_id", b.SQL())
}

func TestJoinBuilder_Reset(t *testing.T) {
	b := NewJoinBuilder(context.Background())
	b.JOINOn("LEFT JOIN posts ON posts.user_id = users.id AND posts.tenant_id = ?", "tenant-X")
	assert.NotEmpty(t, b.SQL())
	assert.Len(t, b.Values(), 1)

	b.Reset(context.Background())
	assert.Empty(t, b.SQL())
	assert.Empty(t, b.Values())
}
