package filter

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/insei/fmap/v3"
	"github.com/insei/gerpo/types"
)

type QOperation[TModel any] interface {
	EQ(val any) ANDOR[TModel]
	NEQ(val any) ANDOR[TModel]
	CT(val any) ANDOR[TModel]
	NCT(val any) ANDOR[TModel]
	BW(val any) ANDOR[TModel]
	NBW(val any) ANDOR[TModel]
	EW(val any) ANDOR[TModel]
	NEW(val any) ANDOR[TModel]
	GT(val any) ANDOR[TModel]
	GTE(val any) ANDOR[TModel]
	LT(val any) ANDOR[TModel]
	LTE(val any) ANDOR[TModel]
}

type Target[TModel any] interface {
	Field(fieldPtr any) QOperation[TModel]
	Group(func(t Target[TModel])) ANDOR[TModel]
	//ParseFromFilters(filters string) error
}

type QueryBuilder[TModel any] interface {
	Field(fieldPtr any) QOperation[TModel]
	Group(func(t Target[TModel])) ANDOR[TModel]
	ToSQL() (string, []any)
}

type ANDOR[TModel any] interface {
	OR() Target[TModel]
	AND() Target[TModel]
}

type baseQuery[TModel any] struct {
	model   *TModel
	columns *types.ColumnsStorage
	ctx     context.Context
	fields  *fieldsStorage
	query   string
	values  []any
}

type fieldsStorage struct {
	fmap.Storage
	jsonFields map[string]fmap.Field
}

func newFieldsStorage(fields fmap.Storage) *fieldsStorage {
	storage := &fieldsStorage{
		Storage:    fields,
		jsonFields: make(map[string]fmap.Field),
	}
	for _, path := range fields.GetAllPaths() {
		f := fields.MustFind(path)
		jsonPath := f.GetTagPath("json", true)
		if jsonPath != "" {
			storage.jsonFields[jsonPath] = f
		}
	}
	return storage
}

func (s *fieldsStorage) GetFieldByJsonTag(tag string) (fmap.Field, bool) {
	f, ok := s.jsonFields[tag]
	return f, ok
}

func NewQueryBuilder[TModel any](model *TModel, fields fmap.Storage, columns *types.ColumnsStorage,
	ctx context.Context) QueryBuilder[TModel] {
	return &baseQuery[TModel]{
		model:   model,
		columns: columns,
		ctx:     ctx,
		fields:  newFieldsStorage(fields),
		query:   "",
		values:  nil,
	}
}

func (q *baseQuery[TModel]) ToSQL() (string, []any) {
	return q.query, q.values
}

func (q *baseQuery[TModel]) Group(f func(t Target[TModel])) ANDOR[TModel] {
	q.query += "("
	f(q)
	q.query += ")"
	return q
}

type QueryOperation[TModel any] func(operation types.Operation, val any) ANDOR[TModel]

func (o QueryOperation[TModel]) EQ(val any) ANDOR[TModel] {
	return o(types.OperationEQ, val)
}

func (o QueryOperation[TModel]) NEQ(val any) ANDOR[TModel] {
	return o(types.OperationNEQ, val)
}

func (o QueryOperation[TModel]) CT(val any) ANDOR[TModel] {
	return o(types.OperationCT, val)
}
func (o QueryOperation[TModel]) NCT(val any) ANDOR[TModel] {
	return o(types.OperationNCT, val)
}

func (o QueryOperation[TModel]) GT(val any) ANDOR[TModel] {
	return o(types.OperationGT, val)
}

func (o QueryOperation[TModel]) GTE(val any) ANDOR[TModel] {
	return o(types.OperationGTE, val)
}

func (o QueryOperation[TModel]) LT(val any) ANDOR[TModel] {
	return o(types.OperationLT, val)
}

func (o QueryOperation[TModel]) LTE(val any) ANDOR[TModel] {
	return o(types.OperationLTE, val)
}

func (o QueryOperation[TModel]) BW(val any) ANDOR[TModel] {
	return o(types.OperationBW, val)
}

func (o QueryOperation[TModel]) NBW(val any) ANDOR[TModel] {
	return o(types.OperationNBW, val)
}

func (o QueryOperation[TModel]) EW(val any) ANDOR[TModel] {
	return o(types.OperationEW, val)
}

func (o QueryOperation[TModel]) NEW(val any) ANDOR[TModel] {
	return o(types.OperationNEW, val)
}

func (q *baseQuery[TModel]) AND() Target[TModel] {
	q.query += " AND "
	return q
}

func (q *baseQuery[TModel]) OR() Target[TModel] {
	q.query += " OR "
	return q
}

func (q *baseQuery[TModel]) appendCondition(cl types.Column, operation types.Operation, val any) {
	filterFn, ok := cl.GetFilterFn(operation)
	if !ok {
		panic(fmt.Errorf("for field %s filter %s option is not available", cl.GetField().GetStructPath(), operation))
	}
	sql, appendValue, err := filterFn(q.ctx, val)
	if err != nil {
		panic(err)
	}
	q.query += sql
	if appendValue {
		q.values = append(q.values, val)
	}
}

func (q *baseQuery[TModel]) Field(fieldPtr any) QOperation[TModel] {
	field, err := q.fields.GetFieldByPtr(q.model, fieldPtr)
	if err != nil {
		panic(err)
	}
	cl, ok := q.columns.Get(field)
	if !ok {
		panic(fmt.Errorf("column for field %s was not found in configured columns", field.GetStructPath()))
	}
	return QueryOperation[TModel](func(operation types.Operation, val any) ANDOR[TModel] {
		q.appendCondition(cl, operation, val)
		return q
	})
}

const filtersRegexpRule = `([a-zA-Z]*):([a-z]{2,3}):([^|}{$\s]+)`

var filtersRegexp = regexp.MustCompile(filtersRegexpRule)

func (q *baseQuery[TModel]) addSimpleCondition(condition string) {
	groups := filtersRegexp.FindAllStringSubmatch(condition, -1)
	if len(groups) > 1 {
		panic(fmt.Errorf("too many groups in condition %s", condition))
	}
	for _, conditionGroup := range groups {
		jsonFieldName := conditionGroup[1]
		opString := conditionGroup[2]
		valueStr := conditionGroup[3]
		field, _ := q.fields.GetFieldByJsonTag(jsonFieldName)
		cl, ok := q.columns.Get(field)
		if !ok {
			panic(fmt.Errorf("column for field %s was not found in configured columns", field.GetStructPath()))
		}
		q.appendCondition(cl, types.Operation(opString), valueStr)
	}
}

func getFirstConditionOperatorIndex(filter string) (int, string) {
	iAND := strings.Index(filter, "$$")
	iOR := strings.Index(filter, "||")
	if iAND > -1 && iAND > iOR {
		return iAND, " AND "
	}
	if iOR > -1 && iOR > iAND {
		return iOR, " OR "
	}
	return -1, ""
}

func (q *baseQuery[TModel]) parseConditions(conditions string) error {
	i := getFirstConditionOperatorIndex(conditions)

}
