package gerpo

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/insei/cast"
	"github.com/insei/fmap/v3"
	"github.com/insei/gerpo/filter"
	"github.com/insei/gerpo/types"
	"github.com/insei/gerpo/virtual"
)

type test struct {
	ID        int        `json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	Name      string     `json:"name"`
	Age       int        `json:"age"`
	Bool      bool       `json:"bool"`
	DeletedAt *time.Time `json:"deleted_at"`
}

const filtersRegexpRule = `([a-zA-Z]*):([a-z]{2,3}):([^|}{$\s]+)`

var filtersRegexp = regexp.MustCompile(filtersRegexpRule)

type fieldsStorage struct {
	fmap.Storage
	jsonFields map[string]fmap.Field
}

func (s *fieldsStorage) GetFieldByJsonTag(tag string) (fmap.Field, bool) {
	f, ok := s.jsonFields[tag]
	return f, ok
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

func checkGroupingSyntax(filters string) error {
	lastOpenGroupIndex := strings.LastIndex(filters, "{")
	firstCloseGroupIndex := strings.Index(filters, "}")
	if (lastOpenGroupIndex > 0 && firstCloseGroupIndex < 0) || (lastOpenGroupIndex < 0 && firstCloseGroupIndex > 0) {
		return fmt.Errorf("filter: syntax grouping error, make sure that grouping is closed and opened with { }")
	}
	if lastOpenGroupIndex > -1 && firstCloseGroupIndex > -1 && lastOpenGroupIndex > firstCloseGroupIndex {
		return fmt.Errorf("filter: syntax grouping error, make sure that grouping is closed and opened with { }")
	}
	if lastOpenGroupIndex > 0 && firstCloseGroupIndex > 0 {
		filters = filters[:lastOpenGroupIndex] + filters[firstCloseGroupIndex+1:]
		return checkGroupingSyntax(filters)
	}
	return nil
}

func checkConditionRegexpGroup(storage *fieldsStorage, conditionGroup []string) error {
	if len(conditionGroup) != 4 {
		if len(conditionGroup) > 0 {
			return fmt.Errorf("incorrect filter: %s", conditionGroup[0])
		}
		return fmt.Errorf("incorrect filter: undeterminated")
	}
	fullcondition := conditionGroup[0]
	jsonFieldName := conditionGroup[1]
	opString := conditionGroup[2]
	valueStr := conditionGroup[3]
	field, ok := storage.GetFieldByJsonTag(jsonFieldName)
	if !ok {
		return fmt.Errorf(" %s field was not found in condition: %s", jsonFieldName, fullcondition)
	}
	if !types.IsSupportedOperation(types.Operation(opString)) {
		return fmt.Errorf("operation %s is not supported in condition: %s", jsonFieldName, fullcondition)
	}
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

func splitGroupFilters(s string) [][]string {
	var all []string
	var conds []string
	strs := strings.Split(s, "$$")
	for _, str := range strs {
		conds = append(conds, "AND")
		ors := strings.Split(str, "||")
		for i := 0; i < len(ors)-1; i++ {
			conds = append(conds, "OR")
		}
		all = append(all, ors...)
	}
	var filters [][]string
	for i, _ := range all {
		filters = append(filters, []string{
			conds[i],
			all[i],
		})
	}
	return filters
}

func ValidateFilters(filters string) error {
	fields, _ := fmap.Get[test]()
	storage := newFieldsStorage(fields)
	groups := filtersRegexp.FindAllStringSubmatch(filters, -1)
	shouldBeEmpty := filters
	for _, conditionGroup := range groups {
		err := checkConditionRegexpGroup(storage, conditionGroup)
		if err != nil {
			return err
		}
		shouldBeEmpty = strings.Replace(shouldBeEmpty, conditionGroup[0], "", 1)
	}
	err := checkGroupingSyntax(filters)
	if err != nil {
		return err
	}
	shouldBeEmpty = replaceAllStrings(shouldBeEmpty, "{", "}", "||", "$$")
	if shouldBeEmpty != "" {
		return fmt.Errorf("incorrect filter: symbols %s at index %d", shouldBeEmpty, strings.Index(filters, shouldBeEmpty))
	}
	return nil
}

func ParseFilters(filters string, b filter.Target[test]) error {
	//fields, _ := fmap.Get[test]()
	//storage := newFieldsStorage(fields)
	firstOpenGroupIndex := strings.Index(filters, "{")
	if firstOpenGroupIndex > -1 && firstOpenGroupIndex != 0 {
		ParseFilters(filters[:firstOpenGroupIndex], b)
	}
	return nil
}

func getFirstConditionOperatorIndex(filter string) int {
	iAND := strings.Index(filter, "$$")
	iOR := strings.Index(filter, "||")
	if iAND > -1 && iAND > iOR {
		return iAND
	}
	if iOR > -1 && iOR > iAND {
		return iOR
	}
	return -1
}

func parse(b filter.Target[test]) {
	condOperatorInd := getFirstConditionOperatorIndex("id:neq:1||id:eq:2||id:eq:3")
	if condOperatorInd > -1 {

	}
}

func TestName(t *testing.T) {
	err := ValidateFilters("id:neq:1||12312312{id:eq:2}||id:eq:3")
	filters := splitGroupFilters("id:neq:1||id:eq:2||id:eq:3")
	_ = filters
	groups := filtersRegexp.FindAllStringSubmatch("id:neq:1||id:eq:2||id:eq:3", -1)
	for _, conditionGroup := range groups {
		jsonFieldName := conditionGroup[1]
		opString := conditionGroup[2]
		valueStr := conditionGroup[3]
		_ = jsonFieldName
		_ = opString
		_ = valueStr
	}
	_ = err
	b, _ := NewBuilder[test]()
	b.Table("test").
		Columns(func(m *test, columns *ColumnBuilder[test]) {
			columns.Column(&m.ID).WithInsertProtection().WithUpdateProtection()
			columns.Column(&m.CreatedAt)
			columns.Column(&m.UpdatedAt)
			columns.Column(&m.Name)
			columns.Column(&m.Age)
			columns.Column(&m.DeletedAt)
			columns.Virtual(&m.Bool).
				WithSQL(func(ctx context.Context) string {
					return ``
				}).
				WithBoolEqFilter(
					func(b *virtual.BoolEQFilterBuilder) {
						b.AddFalseSQLFn(func(ctx context.Context) string { return "test.created_at > now()" })
						b.AddTrueSQLFn(func(ctx context.Context) string { return "test.created_at < now()" })
					})
		}).
		BeforeInsert(func(ctx context.Context, m *test) {
			m.ID = 1
			m.CreatedAt = time.Now()
		}).
		BeforeUpdate(func(ctx context.Context, m *test) {
			updAt := time.Now()
			m.UpdatedAt = &updAt
		})

	repo, err := b.Build()
	repo.GetFirst(context.Background(), func(m *test, b filter.Target[test]) {
		b.
			Field(&m.Name).EQ("Ivan").
			OR().
			Group(func(t filter.Target[test]) {
				t.
					Field(&m.Name).CT("any").
					AND().
					Field(&m.Age).GT(12)
			}).
			AND().
			Field(&m.Bool).EQ(true)
	})
	_ = err
	fmt.Println(repo, err)
}
