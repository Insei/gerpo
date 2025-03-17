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

func TestInsert(t *testing.T) {
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
		WithBeforeInsert(func(ctx context.Context, m *User) {
			id, err := uuid.NewV7()
			if err != nil {
				id = uuid.New()
			}
			m.ID = id
			m.CreatedAt = dateAt
		}).
		Build()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when building repository", err)
	}

	tests := []struct {
		name         string
		model        *User
		setupDb      func(mockDB sqlmock.Sqlmock, dateAt time.Time, m *User)
		repoInsertFn func(repo gerpo.Repository[User], user *User) error
	}{
		{
			name: "Insert",
			model: &User{
				Name: "InsertTest",
			},
			setupDb: func(mockDB sqlmock.Sqlmock, dateAt time.Time, m *User) {
				mockDB.ExpectExec(`INSERT INTO users \(id, created_at, name\) VALUES \(\$1,\$2,\$3\)`).
					WithArgs(sqlmock.AnyArg(), &dateAt, m.Name).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			repoInsertFn: func(repo gerpo.Repository[User], user *User) error {
				return repo.Insert(context.Background(), user)
			},
		},
		{
			name: "Insert with exclude name field",
			model: &User{
				Name: "InsertTest",
			},
			setupDb: func(mockDB sqlmock.Sqlmock, dateAt time.Time, m *User) {
				mockDB.ExpectExec(`INSERT INTO users \(id, created_at\) VALUES \(\$1,\$2\)`).
					WithArgs(sqlmock.AnyArg(), &dateAt).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			repoInsertFn: func(repo gerpo.Repository[User], user *User) error {
				return repo.Insert(context.Background(), user, func(m *User, h query.InsertHelper[User]) {
					h.Exclude(&m.Name)
				})
			},
		},
		{
			name: "Insert with exclude name and createdAt field",
			model: &User{
				Name: "InsertTest",
			},
			setupDb: func(mockDB sqlmock.Sqlmock, dateAt time.Time, m *User) {
				mockDB.ExpectExec(`INSERT INTO users \(id\) VALUES \(\$1\)`).
					WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			repoInsertFn: func(repo gerpo.Repository[User], user *User) error {
				return repo.Insert(context.Background(), user, func(m *User, h query.InsertHelper[User]) {
					h.Exclude(&m.Name, &m.CreatedAt)
				})
			},
		},
		{
			name: "Insert with exclude already excluded field",
			model: &User{
				Name: "InsertTest",
			},
			setupDb: func(mockDB sqlmock.Sqlmock, dateAt time.Time, m *User) {
				mockDB.ExpectExec(`INSERT INTO users \(id\) VALUES \(\$1\)`).
					WithArgs(sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			repoInsertFn: func(repo gerpo.Repository[User], user *User) error {
				return repo.Insert(context.Background(), user, func(m *User, h query.InsertHelper[User]) {
					h.Exclude(&m.Name, &m.CreatedAt, &m.UpdatedAt)
				})
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupDb(mockDB, dateAt, test.model)

			err := test.repoInsertFn(repo, test.model)
			if err != nil {
				t.Errorf("an error '%s' was not expected when inserting", err)
			}
		})
	}
}
