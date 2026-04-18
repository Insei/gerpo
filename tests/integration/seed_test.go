//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// fixedSeed — детерминированный набор данных для тестов. UUID'ы фиксированные,
// поэтому тесты могут ссылаться на конкретные записи по индексу.
type fixedSeed struct {
	users    []User
	posts    []Post
	comments []Comment
}

// defaultSeed вставляет standard набор: 10 пользователей, 30 постов (по 3 на юзера),
// 50 комментариев (распределены неравномерно между постами и авторами).
// Вызывать после truncateAll.
func defaultSeed(t *testing.T) fixedSeed {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	base := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	seed := fixedSeed{}

	for i := 0; i < 10; i++ {
		u := User{
			ID:        deterministicUUID("user", i),
			Name:      fmt.Sprintf("User %d", i),
			Age:       20 + i,
			CreatedAt: base.Add(time.Duration(i) * time.Hour),
		}
		if i%2 == 0 {
			email := fmt.Sprintf("user%d@example.com", i)
			u.Email = &email
		}
		seed.users = append(seed.users, u)
	}

	for i := 0; i < 30; i++ {
		userIdx := i / 3
		p := Post{
			ID:        deterministicUUID("post", i),
			UserID:    seed.users[userIdx].ID,
			Title:     fmt.Sprintf("Post %d", i),
			Content:   fmt.Sprintf("Content of post %d", i),
			Published: i%2 == 0,
			CreatedAt: base.Add(time.Duration(i) * 30 * time.Minute),
		}
		if p.Published {
			pa := p.CreatedAt.Add(5 * time.Minute)
			p.PublishedAt = &pa
		}
		seed.posts = append(seed.posts, p)
	}

	for i := 0; i < 50; i++ {
		postIdx := i % 30
		userIdx := (i*7 + 3) % 10
		seed.comments = append(seed.comments, Comment{
			ID:        deterministicUUID("comment", i),
			PostID:    seed.posts[postIdx].ID,
			UserID:    seed.users[userIdx].ID,
			Body:      fmt.Sprintf("Comment body %d", i),
			CreatedAt: base.Add(time.Duration(i) * 15 * time.Minute),
		})
	}

	insertSeed(ctx, t, seed)
	return seed
}

func insertSeed(ctx context.Context, t *testing.T, seed fixedSeed) {
	t.Helper()
	tx, err := pgx5Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("seed: begin tx: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, u := range seed.users {
		_, err := tx.Exec(ctx, `INSERT INTO users (id, name, email, age, created_at) VALUES ($1,$2,$3,$4,$5)`,
			u.ID, u.Name, u.Email, u.Age, u.CreatedAt)
		if err != nil {
			t.Fatalf("seed user: %v", err)
		}
	}
	for _, p := range seed.posts {
		_, err := tx.Exec(ctx, `INSERT INTO posts (id, user_id, title, content, published, published_at, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			p.ID, p.UserID, p.Title, p.Content, p.Published, p.PublishedAt, p.CreatedAt)
		if err != nil {
			t.Fatalf("seed post: %v", err)
		}
	}
	for _, c := range seed.comments {
		_, err := tx.Exec(ctx, `INSERT INTO comments (id, post_id, user_id, body, created_at) VALUES ($1,$2,$3,$4,$5)`,
			c.ID, c.PostID, c.UserID, c.Body, c.CreatedAt)
		if err != nil {
			t.Fatalf("seed comment: %v", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("seed: commit: %v", err)
	}
}

// deterministicUUID создаёт предсказуемый UUID от (entity, idx), чтобы тесты могли
// ссылаться на конкретную запись по индексу и оставаться воспроизводимыми.
func deterministicUUID(entity string, idx int) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("%s-%d", entity, idx)))
}

func nowUTC() time.Time {
	return time.Now().UTC()
}
