package sql

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"github.com/google/uuid"
	"github.com/insei/fmap/v3"

	"github.com/insei/gerpo/types"
)

func genEQFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		if value == nil {
			return fmt.Sprintf("%s IS NULL", query), false
		}
		return fmt.Sprintf("%s = ?", query), true
	}
}
func genNEQFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		if value == nil {
			return fmt.Sprintf("%s IS NOT NULL", query), false
		}
		return fmt.Sprintf("%s != ?", query), true
	}
}

func genLTFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return fmt.Sprintf("%s < ?", query), true
	}
}
func genLTEFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return fmt.Sprintf("%s <= ?", query), true
	}
}

func genGTFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return fmt.Sprintf("%s > ?", query), true
	}
}
func genGTEFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return fmt.Sprintf("%s >= ?", query), true
	}
}

func genINFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		fPtr := ((*[2]unsafe.Pointer)(unsafe.Pointer(&value)))[1]
		anyArr := (*[]any)(fPtr)
		if anyArr != nil && len(*anyArr) > 0 && len(*anyArr) < 9000 {
			placeholders := strings.Repeat("?,", len(*anyArr))
			placeholders = placeholders[:len(placeholders)-1]
			return fmt.Sprintf("%s IN (%s)", query, placeholders), true
		}
		return "", false
	}
}
func genNINFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		fPtr := ((*[2]unsafe.Pointer)(unsafe.Pointer(&value)))[1]
		anyArr := (*[]any)(fPtr)
		if anyArr != nil && len(*anyArr) > 0 && len(*anyArr) < 9000 {
			placeholders := strings.Repeat("?,", len(*anyArr))
			placeholders = placeholders[:len(placeholders)-1]
			return fmt.Sprintf("%s NOT IN (%s)", query, placeholders), true
		}
		return "", false
	}
}

func genCTFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return fmt.Sprintf("LOWER(%s)", query) + " LIKE LOWER('%' || ? || '%')", true
	}
}
func genNCTFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return fmt.Sprintf("LOWER(%s)", query) + " NOT LIKE LOWER('%' || ? || '%')", true
	}
}

func genBWFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return fmt.Sprintf("LOWER(%s)", query) + " LIKE LOWER(? || '%')", true
	}
}
func genNBWFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return fmt.Sprintf("LOWER(%s)", query) + " NOT LIKE LOWER(? || '%')", true
	}
}

func genEWFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return fmt.Sprintf("LOWER(%s)", query) + " LIKE LOWER('%' || ?)", true
	}
}

func genNEWFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return fmt.Sprintf("LOWER(%s)", query) + " NOT LIKE LOWER('%' || ?)", true
	}
}

func GetDefaultTypeFilters(field fmap.Field, query string) map[types.Operation]func(ctx context.Context, value any) (string, bool) {
	filters := make(map[types.Operation]func(ctx context.Context, value any) (string, bool))
	if field.GetType().Kind() == reflect.Ptr {
		filters[types.OperationEQ] = genEQFn(query)
		filters[types.OperationNEQ] = genNEQFn(query)
	}

	derefType := field.GetDereferencedType()
	switch derefType.Kind() {
	case reflect.Bool:
		filters[types.OperationEQ] = genEQFn(query)
		filters[types.OperationNEQ] = genNEQFn(query)
	case reflect.String:
		filters[types.OperationEQ] = genEQFn(query)
		filters[types.OperationNEQ] = genNEQFn(query)
		filters[types.OperationIN] = genINFn(query)
		filters[types.OperationNIN] = genNINFn(query)
		filters[types.OperationCT] = genCTFn(query)
		filters[types.OperationNCT] = genNCTFn(query)
		filters[types.OperationBW] = genBWFn(query)
		filters[types.OperationNBW] = genNBWFn(query)
		filters[types.OperationEW] = genEWFn(query)
		filters[types.OperationNEW] = genNEWFn(query)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		filters[types.OperationEQ] = genEQFn(query)
		filters[types.OperationNEQ] = genNEQFn(query)
		filters[types.OperationLT] = genLTFn(query)
		filters[types.OperationLTE] = genLTEFn(query)
		filters[types.OperationGT] = genGTFn(query)
		filters[types.OperationGTE] = genGTEFn(query)
		filters[types.OperationIN] = genINFn(query)
		filters[types.OperationNIN] = genNINFn(query)
	default:
		switch derefType {
		case reflect.TypeOf(time.Time{}):
			filters[types.OperationLT] = genLTFn(query)
			filters[types.OperationGT] = genGTFn(query)
		case reflect.TypeOf(uuid.UUID{}):
			filters[types.OperationEQ] = genEQFn(query)
			filters[types.OperationNEQ] = genNEQFn(query)
			filters[types.OperationIN] = genINFn(query)
			filters[types.OperationNIN] = genNINFn(query)
		}
	}
	return filters
}

type StringWhereBuilder struct {
	ctx    context.Context
	sql    string
	values []any
}

func (b *StringWhereBuilder) SQL() string {
	return b.sql
}

func (b *StringWhereBuilder) Values() []any {
	return b.values
}

func (b *StringWhereBuilder) StartGroup() {
	b.sql += "("
}
func (b *StringWhereBuilder) EndGroup() {
	b.sql += ")"
}

func (b *StringWhereBuilder) AND() {
	b.sql += " AND "
}

func (b *StringWhereBuilder) OR() {
	b.sql += " OR "
}

func (b *StringWhereBuilder) AppendSQLWithValues(sql string, appendValue bool, value any) {
	b.sql += sql
	if appendValue {
		b.values = append(b.values, value)
	}
}

func (b *StringWhereBuilder) appendValue(operation types.Operation, val any) {
	switch values := val.(type) {
	case []any:
		if len(values) > 0 && operation == types.OperationIN || operation == types.OperationNIN {
			firstValTypeOf := reflect.ValueOf(values[0])
			if firstValTypeOf.Kind() == reflect.Slice {
				for i := 0; i < firstValTypeOf.Len(); i++ {
					b.values = append(b.values, firstValTypeOf.Index(i).Interface())
				}
			} else {
				b.values = append(b.values, values...)
			}
		}
	default:
		b.values = append(b.values, val)
	}
}

func (b *StringWhereBuilder) AppendCondition(cl types.Column, operation types.Operation, val any) error {
	filterFn, ok := cl.GetFilterFn(operation)
	if !ok {
		return fmt.Errorf("for field %s whereSQL %s option is not available", cl.GetField().GetStructPath(), operation)
	}

	sql, appendValue, err := filterFn(b.ctx, val)
	if err != nil {
		return err
	}
	b.sql += sql
	if !appendValue {
		return nil
	}
	b.appendValue(operation, val)
	return nil
}
