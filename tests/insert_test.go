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
	repo, err := gerpo.New[User]().
		Adapter(databasesql.NewAdapter(db)).
		Table("users").
		Columns(func(m *User, columns *gerpo.ColumnBuilder[User]) {
			columns.Field(&m.ID).OmitOnUpdate()
			columns.Field(&m.CreatedAt).OmitOnUpdate()
			columns.Field(&m.UpdatedAt).OmitOnInsert()
			columns.Field(&m.Name)
			columns.Field(&m.DeletedAt).OmitOnInsert()
			columns.Field(&m.VirtualString).AsVirtual().
				Compute("convert(varchar(25), getdate(), 120)") //Check that not appends to update sql query
			columns.Field(&m.AnotherTable).WithTable("<another_table>") // in real usage join should be configured in WithQuery, but now we simply drops this column
		}).
		WithBeforeInsert(func(ctx context.Context, m *User) error {
			id, err := uuid.NewV7()
			if err != nil {
				id = uuid.New()
			}
			m.ID = id
			m.CreatedAt = dateAt
			return nil
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
				mockDB.ExpectExec(`INSERT INTO users \(id, created_at, name\) VALUES \(\?,\?,\?\)`).
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
				mockDB.ExpectExec(`INSERT INTO users \(id, created_at\) VALUES \(\?,\?\)`).
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
				mockDB.ExpectExec(`INSERT INTO users \(id\) VALUES \(\?\)`).
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
				mockDB.ExpectExec(`INSERT INTO users \(id\) VALUES \(\?\)`).
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
