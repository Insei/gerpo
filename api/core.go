package api

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/insei/cast"
	"github.com/insei/fmap/v3"
	"github.com/insei/gerpo/types"
)

const filtersRegexpRule = `([0-9a-zA-Z_]*):([a-z]{2,3}):([^|}{$\s]+)`

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

func (c *core) checkConditionRegexpGroup(conditionGroup []string) error {
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
		return fmt.Errorf(" %s field was not found in available filters: %s", jsonFieldName, fullcondition)
	}
	op := types.Operation(opString)
	if !col.IsAvailableFilterOperation(types.Operation(opString)) {
		return fmt.Errorf("operation %s is not supported in condition: %s", jsonFieldName, fullcondition)
	}
	field := col.GetField()
	var err error
	if op == types.OperationIN || op == types.OperationNIN {
		valuesStrs := strings.Split(valueStr, ",")
		for _, valueStr := range valuesStrs {
			_, err := cast.ToReflect(valueStr, col.GetField().GetType())
			if err != nil {
				break
			}
		}
	} else {
		_, err = cast.ToReflect(valueStr, col.GetField().GetType())
	}
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

func (c *core) ValidateFilters(filters string) error {
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

func (c *core) ValidateSorts(sorts string) error {
	sorts = strings.TrimSpace(sorts)
	if len(sorts) == 0 {
		return nil
	}
	sortsArr := strings.Split(sorts, ",")
	for _, jsonSortTag := range sortsArr {
		isASC := strings.HasSuffix(jsonSortTag, "+")
		isDESC := strings.HasSuffix(jsonSortTag, "-")
		if isASC || isDESC {
			jsonSortTag = jsonSortTag[0 : len(jsonSortTag)-1]
		}
		if !slices.Contains(c.availSorts, jsonSortTag) {
			return fmt.Errorf("sort is not available for field %s", jsonSortTag)
		}
	}
	return nil
}

func (c *core) parseCondition(filterCondition string) (types.Column, types.Operation, any, error) {
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
		var value any
		if op == types.OperationIN || op == types.OperationNIN {
			values := []any{}
			valuesStrs := strings.Split(valueStr, ",")
			for _, valueStr := range valuesStrs {
				value, _ := cast.ToReflect(valueStr, col.GetField().GetType())
				values = append(values, value)
			}
			value = values
		} else {
			value, _ = cast.ToReflect(valueStr, col.GetField().GetType())
		}
		return col, op, value, nil
	}
	return nil, "", nil, fmt.Errorf("unexpected parse conditions error")
}

type FieldConnector interface {
	Link(dtoFieldPtr, modelFieldPtr any) error
}

type applier struct {
	factory *core
	opts    []func(b types.ConditionBuilder)
}

type core struct {
	columns      map[string]types.Column
	availFilters map[string][]string
	availSorts   []string
}

func (c *core) GetAvailableFilters() map[string][]string {
	avail := make(map[string][]string)
	for jsonTag, col := range c.columns {
		var operations []string
		operationsTyped := col.GetAvailableFilterOperations()
		for _, op := range operationsTyped {
			operations = append(operations, string(op))
		}
		if len(operations) > 0 {
			avail[jsonTag] = operations
		}
	}
	return avail
}

func (c *core) GetAvailableSorts() []string {
	if c.availSorts != nil {
		return c.availSorts
	}
	avail := make([]string, 0)
	for jsonTag, col := range c.columns {
		if col.IsAllowedAction(types.SQLActionSort) {
			avail = append(avail, jsonTag)
		}
	}
	c.availSorts = avail
	return avail
}

func (c *core) link(col types.Column, dtoField fmap.Field) error {
	key := dtoField.GetTagPath("json", true)
	if key == "" {
		return nil
	}
	operations := col.GetAvailableFilterOperations()
	if len(operations) < 1 {
		return nil
	}
	c.columns[key] = col
	for _, op := range operations {
		c.availFilters[key] = append(c.availFilters[key], string(op))
	}
	return nil
}

func (c *core) linkField(storage *types.ColumnsStorage, dto, model, dtoFieldPtr, modelFieldPtr any) error {
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

func (c *core) initDefaultFields(storage *types.ColumnsStorage, dto, model any) error {
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

func NewAPICore[TModel, TDto any](columns *types.ColumnsStorage, opts ...APIConnectorOption[TModel]) (Core, error) {
	dto := new(TDto)
	model := new(TModel)
	c := &core{
		columns:      make(map[string]types.Column),
		availFilters: make(map[string][]string),
	}
	err := c.initDefaultFields(columns, dto, model)
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		switch optTyped := opt.(type) {
		case *advancedLinksOption[TDto, TModel]:
			optTyped.fn(dto, model, FieldLinkConnectorFn(func(dtoFieldPtr, modelFieldPtr any) error {
				return c.linkField(columns, dto, model, dtoFieldPtr, modelFieldPtr)
			}))
		default:
			optTyped.apply(c)
		}
	}
	return c, nil
}

func (c *core) ApplyFilters(filters string, target types.WhereTarget) {
	for lastInd, i := 0, 0; i < len(filters); i++ {
		switch filters[i : i+1] {
		case "{":
			lastInd = strings.LastIndex(filters[i:], "}") + i
			andor := target.Group(func(t types.WhereTarget) {
				c.ApplyFilters(filters[i+1:lastInd], t)
			})
			lastInd++
			if lastInd >= len(filters)-1 {
				continue
			}
			next := filters[lastInd : lastInd+1]
			if next == "|" {
				andor.OR()
			} else {
				andor.AND()
			}
			lastInd += 2
			i = lastInd
		case "|":
			col, op, val, err := c.parseCondition(filters[lastInd:i])
			if err != nil {
				panic(err)
			}
			target.Column(col).OP(op, val).OR()
			lastInd = i + 2
			i++
		case "$":
			col, op, val, err := c.parseCondition(filters[lastInd:i])
			if err != nil {
				panic(err)
			}
			target.Column(col).OP(op, val).AND()
			lastInd = i + 2
			i++
		default:
			if i == len(filters)-1 && lastInd < len(filters)-1 {
				col, op, val, err := c.parseCondition(filters[lastInd:])
				if err != nil {
					panic(err)
				}
				target.Column(col).OP(op, val)
			}
			continue
		}
	}
}

func (c *core) ApplySorts(sorts string, target types.OrderTarget) {
	sorts = strings.TrimSpace(sorts)
	sortsArr := strings.Split(sorts, ",")
	for _, jsonSortTag := range sortsArr {
		isASC := strings.HasSuffix(jsonSortTag, "+")
		isDESC := strings.HasSuffix(jsonSortTag, "-")
		if isASC || isDESC {
			jsonSortTag = jsonSortTag[0 : len(jsonSortTag)-1]
		}
		col, ok := c.columns[jsonSortTag]
		if !ok {
			continue
		}
		orderCol := target.Column(col)
		if isASC {
			orderCol.ASC()
		} else {
			orderCol.DESC()
		}
	}
}
