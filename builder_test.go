package gerpo

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/insei/gerpo/query"
)

type mockModel struct {
	ID   int
	Name string
}

func TestBuilder_Table(t *testing.T) {
	tests := []struct {
		name      string
		table     string
		wantTable string
	}{
		{"valid_table", "users", "users"},
		{"empty_table", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder[mockModel]().(*builder[mockModel])
			b.Table(tt.table)
			if b.table != tt.wantTable {
				t.Errorf("expected table %v, got %v", tt.wantTable, b.table)
			}
		})
	}
}

func TestBuilder_DB(t *testing.T) {
	mockDB := &sql.DB{}
	tests := []struct {
		name string
		db   *sql.DB
	}{
		{"valid_db", mockDB},
		{"nil_db", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder[mockModel]().(*builder[mockModel])
			b.DB(tt.db)
			if b.db != tt.db {
				t.Errorf("expected db %v, got %v", tt.db, b.db)
			}
		})
	}
}

func TestBuilder_Build(t *testing.T) {
	mockDB := &sql.DB{}
	tests := []struct {
		name        string
		configure   func(b *builder[mockModel])
		expectError error
	}{
		{
			"valid_build",
			func(b *builder[mockModel]) {
				b.DB(mockDB).Table("users").Columns(func(m *mockModel, columns *ColumnBuilder[mockModel]) {
					columns.Column(&m.ID)
				})
			},
			nil,
		},
		{
			"missing_table",
			func(b *builder[mockModel]) {
				b.DB(mockDB)
			},
			errors.New("no table found"),
		},
		{
			"missing_db",
			func(b *builder[mockModel]) {
				b.Table("users")
			},
			errors.New("no database found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder[mockModel]().(*builder[mockModel])
			tt.configure(b)
			_, err := b.Build()
			if (err != nil && tt.expectError == nil) || (err == nil && tt.expectError != nil) || (err != nil && err.Error() != tt.expectError.Error()) {
				t.Errorf("expected error %v, got %v", tt.expectError, err)
			}
		})
	}
}

func TestBuilder_Columns(t *testing.T) {
	mockColumnBuilder := func(m *mockModel, columns *ColumnBuilder[mockModel]) {}
	tests := []struct {
		name               string
		columnBuilderFn    func(*mockModel, *ColumnBuilder[mockModel])
		expectedColumnFunc interface{}
	}{
		{"valid_columns", mockColumnBuilder, mockColumnBuilder},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder[mockModel]().(*builder[mockModel])
			b.Columns(tt.columnBuilderFn)
			if &b.columnBuilderFn == nil || b.columnBuilderFn == nil {
				t.Errorf("columnBuilderFn is not set")
			}
		})
	}
}

func TestBuilder_WithQuery(t *testing.T) {
	mockQueryFn := func(m *mockModel, h query.PersistentHelper[mockModel]) {}
	tests := []struct {
		name     string
		queryFn  func(*mockModel, query.PersistentHelper[mockModel])
		expected int
	}{
		{"append_query", mockQueryFn, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder[mockModel]().(*builder[mockModel])
			b.WithQuery(tt.queryFn)
			if len(b.opts) != tt.expected {
				t.Errorf("expected %v opts, got %v", tt.expected, len(b.opts))
			}
		})
	}
}

func TestBuilder_BeforeInsert(t *testing.T) {
	mockBeforeInsertFn := func(ctx context.Context, m *mockModel) {}
	tests := []struct {
		name           string
		beforeInsertFn func(context.Context, *mockModel)
		expected       int
	}{
		{"valid_before_insert", mockBeforeInsertFn, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder[mockModel]().(*builder[mockModel])
			b.BeforeInsert(tt.beforeInsertFn)
			if len(b.opts) != tt.expected {
				t.Errorf("expected %v opts, got %v", tt.expected, len(b.opts))
			}
		})
	}
}

func TestBuilder_AfterInsert(t *testing.T) {
	mockAfterInsertFn := func(ctx context.Context, m *mockModel) {}
	tests := []struct {
		name          string
		afterInsertFn func(context.Context, *mockModel)
		expected      int
	}{
		{"valid_after_insert", mockAfterInsertFn, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder[mockModel]().(*builder[mockModel])
			b.AfterInsert(tt.afterInsertFn)
			if len(b.opts) != tt.expected {
				t.Errorf("expected %v opts, got %v", tt.expected, len(b.opts))
			}
		})
	}
}

func TestBuilder_WithErrorTransformer(t *testing.T) {
	mockErrorTransformer := func(err error) error {
		return errors.New("transformed error")
	}
	tests := []struct {
		name             string
		errorTransformer func(error) error
		expected         int
	}{
		{"valid_error_transformer", mockErrorTransformer, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBuilder[mockModel]().(*builder[mockModel])
			b.WithErrorTransformer(tt.errorTransformer)
			if len(b.opts) != tt.expected {
				t.Errorf("expected %v opts, got %v", tt.expected, len(b.opts))
			}
		})
	}
}
