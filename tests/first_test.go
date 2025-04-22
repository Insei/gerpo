package tests

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/insei/gerpo"
	"github.com/insei/gerpo/executor/adapters/databasesql"
	"github.com/insei/gerpo/executor/adapters/pgx4"
	"github.com/insei/gerpo/query"
	"github.com/jackc/pgx/v4/pgxpool"
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
				return
			}
			if model.Name != "TestName" {
				t.Errorf("Expected name to be TestName, got %s", model.Name)
			}
		})
	}
}

// BenchmarkGetFirst - Bench for check allocations and performance metrics, with db mock(without implementation).
// cmd: go test -bench=^\QBenchmarkGetFirst\E$ -cpuprofile=cpu-result -memprofile=mem-result -benchmem -count 20
// For now: per operation 2,6kb RAM, 0.000002479s. GC works ok, no memory leaks. All metrics stable.
// cpu: AMD Ryzen 9 5950X 16-Core Processor
// BenchmarkGetFirst/GetFirst-32             461305              2523 ns/op            2584 B/op         47 allocs/op
// BenchmarkGetFirst/GetFirst-32             459384              2489 ns/op            2584 B/op         47 allocs/op
// BenchmarkGetFirst/GetFirst-32             470290              2556 ns/op            2584 B/op         47 allocs/op
// BenchmarkGetFirst/GetFirst-32             481104              2516 ns/op            2584 B/op         47 allocs/op
// BenchmarkGetFirst/GetFirst-32             482695              2488 ns/op            2584 B/op         47 allocs/op
// BenchmarkGetFirst/GetFirst-32             486386              2480 ns/op            2584 B/op         47 allocs/op
// BenchmarkGetFirst/GetFirst-32             479238              2453 ns/op            2584 B/op         47 allocs/op
// BenchmarkGetFirst/GetFirst-32             477159              2484 ns/op            2584 B/op         47 allocs/op
// BenchmarkGetFirst/GetFirst-32             472178              2458 ns/op            2584 B/op         47 allocs/op
// BenchmarkGetFirst/GetFirst-32             477238              2484 ns/op            2584 B/op         47 allocs/op
// BenchmarkGetFirst/GetFirst-32             474735              2451 ns/op            2584 B/op         47 allocs/op
// BenchmarkGetFirst/GetFirst-32             449420              2465 ns/op            2584 B/op         47 allocs/op
// BenchmarkGetFirst/GetFirst-32             489318              2442 ns/op            2584 B/op         47 allocs/op
// BenchmarkGetFirst/GetFirst-32             491554              2513 ns/op            2584 B/op         47 allocs/op
// BenchmarkGetFirst/GetFirst-32             502855              2513 ns/op            2584 B/op         47 allocs/op
// BenchmarkGetFirst/GetFirst-32             472483              2467 ns/op            2584 B/op         47 allocs/op
// BenchmarkGetFirst/GetFirst-32             503473              2498 ns/op            2584 B/op         47 allocs/op
// BenchmarkGetFirst/GetFirst-32             462600              2495 ns/op            2584 B/op         47 allocs/op
// BenchmarkGetFirst/GetFirst-32             439519              2502 ns/op            2584 B/op         47 allocs/op
// BenchmarkGetFirst/GetFirst-32             483416              2483 ns/op            2584 B/op         47 allocs/op
func BenchmarkGetFirst(b *testing.B) {
	type User struct {
		ID        uuid.UUID
		CreatedAt time.Time
		UpdatedAt *time.Time
		Name      string
		DeletedAt *time.Time
		//VirtualString string
		//LastLoginTime *time.Time
	}
	dateAt := time.Now().UTC()
	db, err := pgxpool.Connect(context.Background(), "postgresql://postgres:903632as@localhost:5432/gerpo_test?sslmode=disable")
	adapter := pgx4.NewPoolAdapter(db)
	//db := newMockDB()
	//db.QueryContextFn = func(ctx context.Context, query string, args ...any) (extypes.Rows, error) {
	//	return &mockRows{alwaysNext: true}, nil
	//}
	repo, err := gerpo.NewBuilder[User]().
		DB(adapter).
		Table("users").
		Columns(func(m *User, columns *gerpo.ColumnBuilder[User]) {
			columns.Field(&m.ID).AsColumn().WithUpdateProtection()
			columns.Field(&m.CreatedAt).AsColumn().WithUpdateProtection()
			columns.Field(&m.UpdatedAt).AsColumn().WithInsertProtection()
			columns.Field(&m.Name).AsColumn()
			columns.Field(&m.DeletedAt).AsColumn().WithInsertProtection()
			//columns.Field(&m.VirtualString).AsVirtual().
			//	WithSQL(func(ctx context.Context) string {
			//		return `convert(varchar(25), getdate(), 120)`
			//	}) //Check that not appends to update sql query
			//columns.Field(&m.LastLoginTime).AsVirtual().WithSQL(func(ctx context.Context) string {
			//	return "MAX(logins.created_at)"
			//})
		}).
		WithSoftDeletion(func(m *User, softDeletion *gerpo.SoftDeletionBuilder[User]) {
			softDeletion.Field(&m.DeletedAt).SetValueFn(func(ctx context.Context) any {
				return &dateAt
			})
		}).
		WithQuery(func(m *User, h query.PersistentHelper[User]) {
			//h.LeftJoin(func(ctx context.Context) string {
			//	return `logins ON logins.user_id = users.id`
			//})
			h.Where().Field(&m.DeletedAt).EQ(nil)
		}).
		Build()
	if err != nil {
		panic(err)
	}
	//id := uuid.New()
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
					//h.Where().Field(&m.ID).EQ(id)
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
					b.Errorf("an error '%s' was not expected when get first", err)
				}
			}
		})
	}
}

func BenchmarkGetOneFromDb(b *testing.B) {
	type User struct {
		ID        uuid.UUID
		CreatedAt time.Time
		UpdatedAt *time.Time
		Name      string
		DeletedAt *time.Time
		//VirtualString string
		//LastLoginTime *time.Time
	}
	db, err := pgxpool.Connect(context.Background(), "postgresql://postgres:903632as@localhost:5432/gerpo_test?sslmode=disable")
	if err != nil {
		panic(err)
	}
	adapter := pgx4.NewPoolAdapter(db)
	for i := 0; i < b.N; i++ {
		rows, err := adapter.QueryContext(context.Background(), "SELECT users.id, users.created_at, users.updated_at, users.name, users.deleted_at FROM users WHERE (users.deleted_at IS NULL) ORDER BY users.created_at ASC LIMIT 1")
		if err != nil {
			panic(err)
		}
		user := new(User)
		rows.Next()
		err = rows.Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt, &user.Name, &user.DeletedAt)
		if err != nil {
			panic(err)
		}
		rows.Close()
	}
}
