package query

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/insei/cast"
	"github.com/insei/fmap/v3"
	"github.com/insei/gerpo/sql"
	"github.com/insei/gerpo/types"
)

const filtersRegexpRule = `([a-zA-Z]*):([a-z]{2,3}):([^|}{$\s]+)`

var filtersRegexp = regexp.MustCompile(filtersRegexpRule)

func checkGroupingSyntax(filters string) error {
	lastOpenGroupIndex := strings.LastIndex(filters, "{")
	firstCloseGroupIndex := strings.Index(filters, "}")
	if (lastOpenGroupIndex > 0 && firstCloseGroupIndex < 0) || (lastOpenGroupIndex < 0 && firstCloseGroupIndex > 0) {
		return fmt.Errorf("whereSQL: syntax grouping error, make sure that grouping is closed and opened with { }")
	}
	if lastOpenGroupIndex > -1 && firstCloseGroupIndex > -1 && lastOpenGroupIndex > firstCloseGroupIndex {
		return fmt.Errorf("whereSQL: syntax grouping error, make sure that grouping is closed and opened with { }")
	}
	if lastOpenGroupIndex > 0 && firstCloseGroupIndex > 0 {
		filters = filters[:lastOpenGroupIndex] + filters[firstCloseGroupIndex+1:]
		return checkGroupingSyntax(filters)
	}
	return nil
}

func (c *APIConnectorFactory) checkConditionRegexpGroup(conditionGroup []string) error {
	if len(conditionGroup) != 4 {
		if len(conditionGroup) > 0 {
			return fmt.Errorf("incorrect whereSQL: %s", conditionGroup[0])
		}
		return fmt.Errorf("incorrect whereSQL: undeterminated")
	}
	fullcondition := conditionGroup[0]
	jsonFieldName := conditionGroup[1]
	opString := conditionGroup[2]
	valueStr := conditionGroup[3]
	col, ok := c.columns[jsonFieldName]
	if !ok {
		return fmt.Errorf(" %s field was not found in condition: %s", jsonFieldName, fullcondition)
	}
	if !col.IsAvailableFilterOperation(types.Operation(opString)) {
		return fmt.Errorf("operation %s is not supported in condition: %s", jsonFieldName, fullcondition)
	}
	field := col.GetField()
	_, err := cast.ToReflect(valueStr, field.GetType())
	if err != nil {
		return fmt.Errorf("can't convert string value for %s field: %w: value - %s, type - %s; in condition: %s",
			jsonFieldName, err, valueStr, field.GetType().String(), fullcondition)
	}
	return nil
}

func replaceAllStrings(str string, subs ...string) string {
	for _, sub := range subs {
		str = strings.ReplaceAll(str, sub, "")
	}
	return str
}

func (c *APIConnectorFactory) ValidateFilters(filters string) error {
	groups := filtersRegexp.FindAllStringSubmatch(filters, -1)
	shouldBeEmpty := filters // At the end this string should be empty, we cut all valid params from this string
	for _, conditionGroup := range groups {
		err := c.checkConditionRegexpGroup(conditionGroup)
		if err != nil {
			return err
		}
		// Cleanup valid whereSQL
		shouldBeEmpty = strings.Replace(shouldBeEmpty, conditionGroup[0], "", 1)
	}
	err := checkGroupingSyntax(filters)
	if err != nil {
		return err
	}
	// Cleanup valid syntax
	shouldBeEmpty = replaceAllStrings(shouldBeEmpty, "{", "}", "||", "$$")
	if shouldBeEmpty != "" {
		return fmt.Errorf("incorrect whereSQL: symbols %s at index %d", shouldBeEmpty, strings.Index(filters, shouldBeEmpty))
	}
	return nil
}

func (c *APIConnectorFactory) parseCondition(filterCondition string) (types.Column, types.Operation, any, error) {
	conditionRegexGroups := filtersRegexp.FindAllStringSubmatch(filterCondition, -1)
	for _, conditionRegexpParts := range conditionRegexGroups {
		jsonFieldName := conditionRegexpParts[1]
		opString := conditionRegexpParts[2]
		valueStr := conditionRegexpParts[3]
		col, ok := c.columns[jsonFieldName]
		if !ok {
			return nil, "", nil, fmt.Errorf("%s: \"%s\" field was not found in columns", filterCondition, jsonFieldName)
		}
		op := types.Operation(opString)
		value, _ := cast.ToReflect(valueStr, col.GetField().GetType())
		return col, op, value, nil
	}
	return nil, "", nil, fmt.Errorf("unexpected parse conditions error")
}

type FieldConnector interface {
	Link(dtoFieldPtr, modelFieldPtr any) error
}

type APIConnector struct {
	factory *APIConnectorFactory
	opts    []func(b *sql.StringWhereBuilder)
}

type APIConnectorFactory struct {
	columns map[string]types.Column
	allowed map[string][]string
}

func (c *APIConnectorFactory) link(col types.Column, dtoField fmap.Field) error {
	key := dtoField.GetTagPath("json", true)
	if key == "" {
		return nil
	}
	allowed := col.GetAllowedActions()
	if len(allowed) < 1 {
		return nil
	}
	c.columns[key] = col
	for _, action := range allowed {
		c.allowed[key] = append(c.allowed[key], string(action))
	}
	return nil
}

func (c *APIConnectorFactory) linkField(storage *types.ColumnsStorage, dto, model, dtoFieldPtr, modelFieldPtr any) error {
	col, err := storage.GetByFieldPtr(model, modelFieldPtr)
	if err != nil {
		return err
	}
	dtoFields, err := fmap.GetFrom(dto)
	if err != nil {
		return err
	}
	dtoField, err := dtoFields.GetFieldByPtr(dto, dtoFieldPtr)
	if err != nil {
		return err
	}
	return c.link(col, dtoField)
}

func (c *APIConnectorFactory) initDefaultFields(storage *types.ColumnsStorage, dto, model any) error {
	dtoFields, err := fmap.GetFrom(dto)
	if err != nil {
		return err
	}
	modelFields, err := fmap.GetFrom(model)
	if err != nil {
		return err
	}
	for _, fieldPath := range dtoFields.GetAllPaths() {
		modelField, ok := modelFields.Find(fieldPath)
		if !ok {
			continue
		}
		col, ok := storage.Get(modelField)
		if !ok {
			continue
		}
		dtoField := dtoFields.MustFind(fieldPath)
		err := c.link(col, dtoField)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *APIConnectorFactory) New() *APIConnector {
	return &APIConnector{
		factory: c,
	}
}

type FieldLinkConnectorFn func(dtoFieldPtr, modelFieldPtr any) error

func (f FieldLinkConnectorFn) Link(dtoFieldPtr, modelFieldPtr any) error {
	return f(dtoFieldPtr, modelFieldPtr)
}

type advancedLinksOption[TDto, TModel any] struct {
	fn func(dto *TDto, model *TModel, conn FieldConnector)
}

func (o *advancedLinksOption[TDto, TModel]) apply(_ *APIConnectorFactory) {

}

type APIConnectorOption[TModel any] interface {
	apply(c *APIConnectorFactory)
}

func WithAdvancedFieldLink[TDto, TModel any](fn func(dto *TDto, model *TModel, conn FieldConnector)) APIConnectorOption[TModel] {
	return &advancedLinksOption[TDto, TModel]{fn: fn}
}

func NewAPIConnectorFactory[TDto, TModel any](storage *types.ColumnsStorage, opts ...APIConnectorOption[TModel]) (*APIConnectorFactory, error) {
	dto := new(TDto)
	model := new(TModel)
	c := &APIConnectorFactory{
		columns: make(map[string]types.Column),
		allowed: make(map[string][]string),
	}
	err := c.initDefaultFields(storage, dto, model)
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		switch optTyped := opt.(type) {
		case *advancedLinksOption[TDto, TModel]:
			optTyped.fn(dto, model, FieldLinkConnectorFn(func(dtoFieldPtr, modelFieldPtr any) error {
				return c.linkField(storage, dto, model, dtoFieldPtr, modelFieldPtr)
			}))
		default:
			optTyped.apply(c)
		}
	}
	return c, nil
}

func (c *APIConnector) AppendFilters(filters string) *APIConnector {
	if strings.TrimSpace(filters) == "" {
		return c
	}
	c.opts = append(c.opts, func(b *sql.StringWhereBuilder) {
		b.AND()
		b.StartGroup()
		for lastInd, i := 0, 0; i < len(filters); i++ {
			switch filters[i : i+1] {
			case "{":
				b.StartGroup()
				lastInd++
			case "}":
				b.EndGroup()
				lastInd++
			case "|":
				col, op, val, err := c.factory.parseCondition(filters[lastInd:i])
				if err != nil {
					panic(err)
				}
				err = b.AppendCondition(col, op, val)
				if err != nil {
					panic(err)
				}
				b.OR()
				lastInd = i + 2
				i++
			case "$":
				col, op, val, err := c.factory.parseCondition(filters[lastInd:i])
				if err != nil {
					panic(err)
				}
				err = b.AppendCondition(col, op, val)
				if err != nil {
					panic(err)
				}
				b.AND()
				lastInd = i + 2
				i++
			default:
				if i == len(filters)-1 && lastInd < len(filters)-1 {
					col, op, val, err := c.factory.parseCondition(filters[lastInd:])
					if err != nil {
						panic(err)
					}
					err = b.AppendCondition(col, op, val)
					if err != nil {
						panic(err)
					}
				}
				continue
			}
		}
		b.EndGroup()
	})
	return c
}

func (c *APIConnector) ApplyWhere(b *sql.StringWhereBuilder) {
	for _, op := range c.opts {
		op(b)
	}
}
