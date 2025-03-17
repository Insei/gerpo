package tests

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/insei/gerpo"
	"github.com/insei/gerpo/executor/adapters/databasesql"
	extypes "github.com/insei/gerpo/executor/types"
	"github.com/insei/gerpo/query"
)

func TestGetFirst(t *testing.T) {
	type User struct {
		ID            uuid.UUID
		CreatedAt     time.Time
		UpdatedAt     *time.Time
		Name          string
		DeletedAt     *time.Time
		VirtualString string
		LastLoginTime *time.Time
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
			columns.Field(&m.LastLoginTime).AsVirtual().WithSQL(func(ctx context.Context) string {
				return "MAX(logins.created_at)"
			})
		}).
		WithSoftDeletion(func(m *User, softDeletion *gerpo.SoftDeletionBuilder[User]) {
			softDeletion.Field(&m.DeletedAt).SetValueFn(func(ctx context.Context) any {
				return &dateAt
			})
		}).
		WithQuery(func(m *User, h query.PersistentHelper[User]) {
			h.LeftJoin(func(ctx context.Context) string {
				return `logins ON logins.user_id = users.id`
			})
			h.Where().Field(&m.DeletedAt).EQ(nil) // Check soft deletion where appending
		}).
		Build()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when building repository", err)
	}
	tests := []struct {
		name           string
		model          *User
		setupDb        func(mockDB sqlmock.Sqlmock, dateAt time.Time, m *User)
		repoGetFirstFn func(repo gerpo.Repository[User]) (*User, error)
	}{
		{
			name: "GetFirst",
			model: &User{
				ID: uuid.New(),
			},
			setupDb: func(mockDB sqlmock.Sqlmock, dateAt time.Time, m *User) {

				mockDB.ExpectQuery(`SELECT users.id, users.created_at, users.updated_at, users.name, users.deleted_at, convert\(varchar\(25\), getdate\(\), 120\), MAX\(logins.created_at\)
						FROM users 
					    LEFT JOIN logins ON logins.user_id = users.id WHERE \(users.deleted_at IS NULL\) ORDER BY users.created_at ASC LIMIT 1`).
					WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "virtual_string", "logins_created_at", "deleted_at"}).
						AddRow(m.ID, dateAt, &dateAt, "TestName", &dateAt, time.Now().String(), &dateAt)).
					RowsWillBeClosed()
			},
			repoGetFirstFn: func(repo gerpo.Repository[User]) (*User, error) {
				return repo.GetFirst(context.Background(), func(m *User, h query.GetFirstHelper[User]) {
					h.OrderBy().Field(&m.CreatedAt).ASC()
				})
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupDb(mockDB, dateAt, test.model)

			model, err := test.repoGetFirstFn(repo)
			if err != nil {
				t.Errorf("an error '%s' was not expected when inserting", err)
			}
			if model.Name != "TestName" {
				t.Errorf("Expected name to be TestName, got %s", model.Name)
			}
		})
	}
}

// BenchmarkGetFirst - Bench for check allocations and performance metrics, with db mock(without implementation).
// cmd: go test -bench=^\QBenchmarkGetFirst\E$ -cpuprofile=cpu-result -memprofile=mem-result -benchmem -count 20
// For now: per operation 4kb RAM, 0.000003379s. GC works ok, no memory leaks. All metrics stable.
// cpu: AMD Ryzen 9 5950X 16-Core Processor
// BenchmarkGetFirst/GetFirst-32             327801              3486 ns/op            4108 B/op         66 allocs/op
// BenchmarkGetFirst/GetFirst-32             343087              3645 ns/op            4108 B/op         66 allocs/op
// BenchmarkGetFirst/GetFirst-32             346983              3446 ns/op            4107 B/op         66 allocs/op
// BenchmarkGetFirst/GetFirst-32             361726              3398 ns/op            4107 B/op         66 allocs/op
// BenchmarkGetFirst/GetFirst-32             342397              3452 ns/op            4107 B/op         66 allocs/op
// BenchmarkGetFirst/GetFirst-32             345679              3406 ns/op            4107 B/op         66 allocs/op
// BenchmarkGetFirst/GetFirst-32             335836              3395 ns/op            4107 B/op         66 allocs/op
// BenchmarkGetFirst/GetFirst-32             356157              3388 ns/op            4107 B/op         66 allocs/op
// BenchmarkGetFirst/GetFirst-32             356085              3395 ns/op            4107 B/op         66 allocs/op
// BenchmarkGetFirst/GetFirst-32             348438              3375 ns/op            4106 B/op         66 allocs/op
// BenchmarkGetFirst/GetFirst-32             360686              3386 ns/op            4106 B/op         66 allocs/op
// BenchmarkGetFirst/GetFirst-32             344098              3396 ns/op            4106 B/op         66 allocs/op
// BenchmarkGetFirst/GetFirst-32             355010              3440 ns/op            4106 B/op         66 allocs/op
// BenchmarkGetFirst/GetFirst-32             342636              3382 ns/op            4106 B/op         66 allocs/op
// BenchmarkGetFirst/GetFirst-32             367304              3364 ns/op            4106 B/op         66 allocs/op
// BenchmarkGetFirst/GetFirst-32             357259              3360 ns/op            4106 B/op         66 allocs/op
// BenchmarkGetFirst/GetFirst-32             359004              3368 ns/op            4106 B/op         66 allocs/op
// BenchmarkGetFirst/GetFirst-32             351720              3368 ns/op            4106 B/op         66 allocs/op
// BenchmarkGetFirst/GetFirst-32             348252              3379 ns/op            4106 B/op         66 allocs/op
// BenchmarkGetFirst/GetFirst-32             361478              3346 ns/op            4106 B/op         66 allocs/op
// BenchmarkGetFirst/GetFirst-32             348252              3379 ns/op            4106 B/op         66 allocs/op
func BenchmarkGetFirst(b *testing.B) {
	type User struct {
		ID            uuid.UUID
		CreatedAt     time.Time
		UpdatedAt     *time.Time
		Name          string
		DeletedAt     *time.Time
		VirtualString string
		LastLoginTime *time.Time
	}
	dateAt := time.Now().UTC()
	db := newMockDB()
	db.QueryContextFn = func(ctx context.Context, query string, args ...any) (extypes.Rows, error) {
		return &mockRows{alwaysNext: true}, nil
	}
	repo, err := gerpo.NewBuilder[User]().
		DB(db).
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
			columns.Field(&m.LastLoginTime).AsVirtual().WithSQL(func(ctx context.Context) string {
				return "MAX(logins.created_at)"
			})
		}).
		WithSoftDeletion(func(m *User, softDeletion *gerpo.SoftDeletionBuilder[User]) {
			softDeletion.Field(&m.DeletedAt).SetValueFn(func(ctx context.Context) any {
				return &dateAt
			})
		}).
		WithQuery(func(m *User, h query.PersistentHelper[User]) {
			h.LeftJoin(func(ctx context.Context) string {
				return `logins ON logins.user_id = users.id`
			})
			h.Where().Field(&m.DeletedAt).EQ(nil)
		}).
		Build()
	if err != nil {
		panic(err)
	}
	id := uuid.New()
	//prepareElapsed := time.Since(dateAt)
	//fmt.Printf("Prepare took %d\n", prepareElapsed.Milliseconds())
	tests := []struct {
		name           string
		model          *User
		repoGetFirstFn func(repo gerpo.Repository[User]) (*User, error)
	}{
		{
			name: "GetFirst",
			model: &User{
				ID: uuid.New(),
			},
			repoGetFirstFn: func(repo gerpo.Repository[User]) (*User, error) {
				return repo.GetFirst(context.Background(), func(m *User, h query.GetFirstHelper[User]) {
					h.Where().Field(&m.ID).EQ(id)
					h.OrderBy().Field(&m.CreatedAt).ASC()
				})
			},
		},
	}
	for _, test := range tests {
		b.Run(test.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := test.repoGetFirstFn(repo)
				if err != nil {
					b.Errorf("an error '%s' was not expected when inserting", err)
				}
			}
		})
	}
}
