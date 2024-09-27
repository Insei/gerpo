package config

import (
	"context"
	"database/sql"

	"github.com/insei/fmap/v3"
	"github.com/insei/gerpo/column"
	"github.com/insei/gerpo/types"
	"github.com/insei/gerpo/virtual"
)

type Procedure uint

const (
	// Insert represents the constant value for the insert operation in the repository.
	Insert = Procedure(iota)
	// Update represents the constant value for the update operation in the repository.
	Update
)

// Column represents a database column in a repository configuration.
//
// Table is the name of the table in the database.
// Name is the name of the column in the database.
// SQLColumn is the SQL representation of the column.
// SQLColumnArgs is a function that returns additional arguments for the SQL column.
// FieldMapF is a function that maps the column value to the field of an entity.
// Protection is a map that defines which operations are protected for this column.
type Column[TEntity any] struct {
	Table         string
	Name          string
	SQLColumn     string
	SQLColumnArgs func(ctx context.Context) []any
	FieldMapF     func(ent *TEntity) any
	Protection    map[Procedure]bool
}

type repositoryConfig[TEntity any] struct {
	Columns []Column[TEntity]
	//JoinBuilder      *JoinBuilder
	Table            string
	GroupBys         string
	BeforeCreateF    []func(context.Context, *TEntity)
	BeforeUpdateF    []func(context.Context, *TEntity)
	SoftDeleteF      map[string]func(context.Context) any
	PersistentFilter func(context.Context, types.IQueryBuilder)
	DB               *sql.DB
	Driver           string
	//Logger           *zap.Logger
	Debug bool

	//placeholderFormat sq.PlaceholderFormat
}

type ColumnBuilder interface {
	Build() types.Column
}

type columnBuilder[TEntity any] struct {
	model    *TEntity
	builders []ColumnBuilder
	fields   fmap.Storage
}

func (b *columnBuilder[TEntity]) getFmapField(fieldFn func(m *TEntity) any) fmap.Field {
	fieldPtr := fieldFn(b.model)
	field, err := b.fields.GetFieldByPtr(b.model, fieldPtr)
	if err != nil {
		panic(err)
	}
	return field
}

func (b *columnBuilder[TEntity]) Column(fieldFn func(m *TEntity) any) *column.Builder {
	field := b.getFmapField(fieldFn)
	builder := column.NewBuilder(field)
	b.builders = append(b.builders, builder)
	return builder
}

func (b *columnBuilder[TEntity]) Virtual(fieldFn func(m *TEntity) any) *virtual.Builder {
	field := b.getFmapField(fieldFn)
	builder := virtual.NewBuilder(field)
	b.builders = append(b.builders, builder)
	return builder
}
func (b *columnBuilder[TEntity]) Build() []types.Column {
	columns := make([]types.Column, 0)
	for _, builder := range b.builders {
		columns = append(columns, builder.Build())
	}
	return columns
}

type test struct {
	Name string
	Bool bool
}
