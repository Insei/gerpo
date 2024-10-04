package gerpo

import (
	"github.com/insei/fmap/v3"
)

// type Procedure uint
//
// const (
//
//	// Insert represents the constant value for the insert operation in the repository.
//	Insert = Procedure(iota)
//	// Update represents the constant value for the update operation in the repository.
//	Update
//
// )
//
// // Column represents a database column in a repository configuration.
// //
// // Table is the name of the table in the database.
// // Name is the name of the column in the database.
// // SQLColumn is the SQL representation of the column.
// // SQLColumnArgs is a function that returns additional arguments for the SQL column.
// // FieldMapF is a function that maps the column value to the field of an entity.
// // Protection is a map that defines which operations are protected for this column.
//
//	type Column[TEntity any] struct {
//		Table         string
//		Name          string
//		SQLColumn     string
//		SQLColumnArgs func(ctx context.Context) []any
//		FieldMapF     func(ent *TEntity) any
//		Protection    map[Procedure]bool
//	}
//
//	type repositoryConfig[TEntity any] struct {
//		Columns []types.Column
//		//JoinBuilder      *JoinBuilder
//		Table            string
//		GroupBys         string
//		BeforeCreateF    []func(context.Context, *TEntity)
//		BeforeUpdateF    []func(context.Context, *TEntity)
//		SoftDeleteF      map[string]func(context.Context) any
//		PersistentFilter func(context.Context, types.IQueryBuilder)
//		DB               *sql.DB
//		Driver           string
//		//Logger           *zap.Logger
//		Debug bool
//
//		//placeholderFormat sq.PlaceholderFormat
//	}
//
//type QueryOperation[TModel any] interface {
//	EQ(val any) ANDOR[TModel]
//	NEQ(val any) ANDOR[TModel]
//	CT(val string) ANDOR[TModel]
//	NCT(val string) ANDOR[TModel]
//	GT(val any) ANDOR[TModel]
//	GTE(val any) ANDOR[TModel]
//}
//
//type QueryTarget[TModel any] interface {
//	Field(fieldPtr any) QueryOperation[TModel]
//	Group(func(t QueryTarget[TModel])) ANDOR[TModel]
//}
//
//type ANDOR[TModel any] interface {
//	OR() QueryTarget[TModel]
//	AND() QueryTarget[TModel]
//}
//
//func testFn() {
//	var tt QueryTarget[test]
//	m := &test{}
//	tt.
//		Group(func(t QueryTarget[test]) {
//			t.
//				Field(&m.Age).GTE(12).
//				OR().
//				Field(&m.ID).EQ(123)
//		}).
//		AND().
//		Group(func(t QueryTarget[test]) {
//			t.Field(&m.Bool).EQ(true)
//		})
//}
//
//type IRepository[TModel any] interface {
//	// GetFirst is a method of the IRepository interface that retrieves the first entity
//	// that matches the given query conditions. It takes a context.Context object and a
//	// queryF function as parameters. The queryF function is used to build the query conditions
//	// using the IQueryBuilder interface. The method returns a pointer to the entity of type T
//	// and an error. If no entity is found or an error occurs during the query execution, nil
//	// is returned for the entity and the error is returned respectively.
//	GetFirst(ctx context.Context, queryF func(builder QueryTarget[TModel])) (*TModel, error)
//
//	// Count returns the number of entities that match the provided query criteria.
//	// It uses the given `ctx` for the context and `queryF` function for building the query.
//	// It returns the count as a `uint64` and an error if any occurred during the query execution.
//	Count(ctx context.Context, queryF func(builder QueryTarget[TModel])) (uint64, error)
//
//	// GetList retrieves a list of entities based on the provided parameters.
//	// It takes a context and a GetListParams object as input and returns a slice of pointers
//	// to the entities and an error. The GetListParams object specifies the page number, page size,
//	// and a query builder function. The query builder function is used to build the SQL query with
//	// various conditions and joins. The method returns a slice of entities that satisfy the query
//	// conditions and an error if any occurred during the retrieval process.
//	GetList(ctx context.Context, params GetListParams) ([]*TModel, error)
//
//	// Insert inserts an entity into the repository.
//	// Parameters:
//	// - ctx: The context.Context for the operation.
//	// - entities: A slice with pointers to the entities to be inserted.
//	// Returns:
//	// - error: An error if the insert operation fails.
//	Insert(ctx context.Context, entities ...*TModel) error
//
//	// Update is a method that updates an entity in the repository based on the provided query and entity.
//	// It takes a context.Context object, a queryF function that is used to build the query, and a pointer to the entity to be updated.
//	// The method returns an error if the update operation fails.
//	// The queryF function takes an IQueryBuilder object to build the query for the update operation.
//	// The entity pointer should point to the object to be updated in the repository.
//	//
//	// Usage:
//	// err := repository.Update(ctx, func(builder IQueryBuilder) {
//	//   builder.Where("id", "=", 1)
//	// }, &entity)
//	//
//	// Params:
//	// - ctx: The context.Context object for the update operation.
//	// - queryF: A function that takes an IQueryBuilder object to build the query.
//	// - entity: A pointer to the entity object to be updated.
//	//
//	// Returns:
//	// - error: An error if the update operation fails.
//	Update(ctx context.Context, queryF func(builder IQueryBuilder[TModel]), entity *TModel) error
//
//	// Delete removes entities from the repository based on the given query function.
//	//
//	// The query function is used to configure the deletion query using the provided IQueryBuilder.
//	//
//	// Example usage:
//	//    repository.Delete(ctx, func(builder IQueryBuilder) {
//	//        builder.Where("id", "=", 1)
//	//    })
//	//
//	// The above code will delete the entity with the id of 1 from the repository.
//	//
//	// Parameters:
//	//   - ctx: The context.Context object for the request.
//	//   - queryF: The function used to configure the deletion query.
//	//             It should accept an IQueryBuilder parameter and configure the query using it.
//	//             The IQueryBuilder interface provides methods for building SQL where conditions.
//	//
//	// Returns:
//	//   - error: An error if the deletion operation fails, nil otherwise.
//	Delete(ctx context.Context, queryF func(builder IQueryBuilder[TModel])) error
//}

func getModelAndFields[TModel any]() (*TModel, fmap.Storage, error) {
	model := new(TModel)
	mustZero(model)
	fields, err := fmap.GetFrom(model)
	if err != nil {
		return nil, nil, err
	}
	return model, fields, nil
}
