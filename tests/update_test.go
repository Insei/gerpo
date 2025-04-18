package tests

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/insei/gerpo"
	"github.com/insei/gerpo/executor/adapters/databasesql"
	"github.com/insei/gerpo/query"
)

func TestUpdate(t *testing.T) {
	type User struct {
		ID            uuid.UUID
		CreatedAt     time.Time
		UpdatedAt     *time.Time
		Name          string
		DeletedAt     *time.Time
		VirtualString string
		AnotherTable  int
	}

	dateAt := time.Now().UTC()
	db, mockDB, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	repo, err := gerpo.NewBuilder[User]().
		DB(databasesql.NewAdapter(db)).
		Table("users").
		Columns(func(m *User, columns *gerpo.ColumnBuilder[User]) {
			columns.Field(&m.ID).AsColumn().WithUpdateProtection()
			columns.Field(&m.CreatedAt).AsColumn().WithUpdateProtection()
			columns.Field(&m.UpdatedAt).AsColumn().WithInsertProtection()
			columns.Field(&m.Name).AsColumn()
			columns.Field(&m.DeletedAt).AsColumn().WithInsertProtection()
			columns.Field(&m.VirtualString).AsVirtual().
				WithSQL(func(ctx context.Context) string {
					return `convert(varchar(25), getdate(), 120)`
				}) //Check that not appends to update sql query
			columns.Field(&m.AnotherTable).AsColumn().WithTable("<another_table>") // in real usage join should be configured in WithQuery, but now we simply drops this column
		}).
		WithSoftDeletion(func(m *User, softDeletion *gerpo.SoftDeletionBuilder[User]) {
			softDeletion.Field(&m.DeletedAt).SetValueFn(func(ctx context.Context) any {
				return &dateAt
			})
		}).
		WithQuery(func(m *User, h query.PersistentHelper[User]) {
			h.Where().Field(&m.DeletedAt).EQ(nil) // Check soft deletion where appending
		}).
		WithBeforeUpdate(func(ctx context.Context, m *User) {
			m.UpdatedAt = &dateAt
		}).
		Build()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when building repository", err)
	}

	tests := []struct {
		name         string
		model        *User
		setupDb      func(mockDB sqlmock.Sqlmock, dateAt time.Time, m *User)
		repoUpdateFn func(repo gerpo.Repository[User], user *User) (int64, error)
	}{
		{
			name: "Update",
			model: &User{
				ID:        uuid.New(),
				Name:      "UpdateTest",
				CreatedAt: time.Now().UTC(),
			},
			setupDb: func(mockDB sqlmock.Sqlmock, dateAt time.Time, m *User) {
				mockDB.ExpectExec(`UPDATE users SET updated_at = \?, name = \?, deleted_at = \? WHERE \(users.deleted_at IS NULL\) AND \(users.id = \?\)`).
					WithArgs(&dateAt, m.Name, m.DeletedAt, m.ID).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			repoUpdateFn: func(repo gerpo.Repository[User], user *User) (int64, error) {
				return repo.Update(context.Background(), user, func(m *User, h query.UpdateHelper[User]) {
					h.Where().Field(&m.ID).EQ(user.ID)
				})
			},
		},
		{
			name: "Update exclude name field",
			model: &User{
				ID:        uuid.New(),
				Name:      "UpdateTest",
				CreatedAt: time.Now().UTC(),
			},
			setupDb: func(mockDB sqlmock.Sqlmock, dateAt time.Time, m *User) {
				mockDB.ExpectExec(`UPDATE users SET updated_at = \?, deleted_at = \? WHERE \(users.deleted_at IS NULL\) AND \(users.id = \?\)`).
					WithArgs(&dateAt, m.DeletedAt, m.ID).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			repoUpdateFn: func(repo gerpo.Repository[User], user *User) (int64, error) {
				return repo.Update(context.Background(), user, func(m *User, h query.UpdateHelper[User]) {
					h.Where().Field(&m.ID).EQ(user.ID)
					h.Exclude(&m.Name)
				})
			},
		},
		{
			name: "Update exclude name, deletedAt fields",
			model: &User{
				ID:        uuid.New(),
				Name:      "UpdateTest",
				CreatedAt: time.Now().UTC(),
			},
			setupDb: func(mockDB sqlmock.Sqlmock, dateAt time.Time, m *User) {
				mockDB.ExpectExec(`UPDATE users SET updated_at = \? WHERE \(users.deleted_at IS NULL\) AND \(users.id = \?\)`).
					WithArgs(&dateAt, m.ID).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			repoUpdateFn: func(repo gerpo.Repository[User], user *User) (int64, error) {
				return repo.Update(context.Background(), user, func(m *User, h query.UpdateHelper[User]) {
					h.Where().Field(&m.ID).EQ(user.ID)
					h.Exclude(&m.Name, &m.DeletedAt)
				})
			},
		},
		{
			name: "Update chained OR where",
			model: &User{
				ID:        uuid.New(),
				Name:      "UpdateTest",
				CreatedAt: time.Now().UTC(),
			},
			setupDb: func(mockDB sqlmock.Sqlmock, dateAt time.Time, m *User) {
				mockDB.ExpectExec(`UPDATE users SET updated_at = \?, name = \?, deleted_at = \? WHERE \(users.deleted_at IS NULL\) AND \(users.id = \?\ OR users.name = \?\)`).
					WithArgs(&dateAt, m.Name, m.DeletedAt, m.ID, m.Name).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			repoUpdateFn: func(repo gerpo.Repository[User], user *User) (int64, error) {
				return repo.Update(context.Background(), user, func(m *User, h query.UpdateHelper[User]) {
					h.Where().
						Field(&m.ID).EQ(user.ID).
						OR().
						Field(&m.Name).EQ(user.Name)
				})
			},
		},
		{
			name: "Update chained AND where",
			model: &User{
				ID:        uuid.New(),
				Name:      "UpdateTest",
				CreatedAt: time.Now().UTC(),
			},
			setupDb: func(mockDB sqlmock.Sqlmock, dateAt time.Time, m *User) {
				mockDB.ExpectExec(`UPDATE users SET updated_at = \?, name = \?, deleted_at = \? WHERE \(users.deleted_at IS NULL\) AND \(users.id = \?\ AND users.name = \?\)`).
					WithArgs(&dateAt, m.Name, m.DeletedAt, m.ID, m.Name).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			repoUpdateFn: func(repo gerpo.Repository[User], user *User) (int64, error) {
				return repo.Update(context.Background(), user, func(m *User, h query.UpdateHelper[User]) {
					h.Where().
						Field(&m.ID).EQ(user.ID).
						AND().
						Field(&m.Name).EQ(user.Name)
				})
			},
		},
		{
			name: "Update not chained where",
			model: &User{
				ID:        uuid.New(),
				Name:      "UpdateTest",
				CreatedAt: time.Now().UTC(),
			},
			setupDb: func(mockDB sqlmock.Sqlmock, dateAt time.Time, m *User) {
				mockDB.ExpectExec(`UPDATE users SET updated_at = \?, name = \?, deleted_at = \? WHERE \(users.deleted_at IS NULL\) AND \(users.id = \?\ AND users.name = \?\)`).
					WithArgs(&dateAt, m.Name, m.DeletedAt, m.ID, m.Name).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			repoUpdateFn: func(repo gerpo.Repository[User], user *User) (int64, error) {
				return repo.Update(context.Background(), user, func(m *User, h query.UpdateHelper[User]) {
					h.Where().Field(&m.ID).EQ(user.ID)
					h.Where().Field(&m.Name).EQ(user.Name)
				})
			},
		},
		// TODO: make update with join support for WHERE clause
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupDb(mockDB, dateAt, test.model)

			_, err := test.repoUpdateFn(repo, test.model)
			if err != nil {
				t.Errorf("an error '%s' was not expected when updating", err)
			}
		})
	}
}
