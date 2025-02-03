package sqlpart

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

func GetFieldTypeFilters(field fmap.Field, sqlColumnString string) map[types.Operation]func(ctx context.Context, value any) (string, bool) {
	filters := make(map[types.Operation]func(ctx context.Context, value any) (string, bool))
	if field.GetType().Kind() == reflect.Ptr {
		filters[types.OperationEQ] = genEQFn(sqlColumnString)
		filters[types.OperationNEQ] = genNEQFn(sqlColumnString)
	}

	derefType := field.GetDereferencedType()
	switch derefType.Kind() {
	case reflect.Bool:
		filters[types.OperationEQ] = genEQFn(sqlColumnString)
		filters[types.OperationNEQ] = genNEQFn(sqlColumnString)
	case reflect.String:
		filters[types.OperationEQ] = genEQFn(sqlColumnString)
		filters[types.OperationNEQ] = genNEQFn(sqlColumnString)
		filters[types.OperationIN] = genINFn(sqlColumnString)
		filters[types.OperationNIN] = genNINFn(sqlColumnString)
		filters[types.OperationCT] = genCTFn(sqlColumnString)
		filters[types.OperationNCT] = genNCTFn(sqlColumnString)
		filters[types.OperationBW] = genBWFn(sqlColumnString)
		filters[types.OperationNBW] = genNBWFn(sqlColumnString)
		filters[types.OperationEW] = genEWFn(sqlColumnString)
		filters[types.OperationNEW] = genNEWFn(sqlColumnString)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		filters[types.OperationEQ] = genEQFn(sqlColumnString)
		filters[types.OperationNEQ] = genNEQFn(sqlColumnString)
		filters[types.OperationLT] = genLTFn(sqlColumnString)
		filters[types.OperationLTE] = genLTEFn(sqlColumnString)
		filters[types.OperationGT] = genGTFn(sqlColumnString)
		filters[types.OperationGTE] = genGTEFn(sqlColumnString)
		filters[types.OperationIN] = genINFn(sqlColumnString)
		filters[types.OperationNIN] = genNINFn(sqlColumnString)
	default:
		switch derefType {
		case reflect.TypeOf(time.Time{}):
			filters[types.OperationLT] = genLTFn(sqlColumnString)
			filters[types.OperationGT] = genGTFn(sqlColumnString)
		case reflect.TypeOf(uuid.UUID{}):
			filters[types.OperationEQ] = genEQFn(sqlColumnString)
			filters[types.OperationNEQ] = genNEQFn(sqlColumnString)
			filters[types.OperationIN] = genINFn(sqlColumnString)
			filters[types.OperationNIN] = genNINFn(sqlColumnString)
		}
	}
	return filters
}

type Where interface {
	StartGroup()
	EndGroup()
	AND()
	OR()
	AppendSQLWithValues(sql string, appendValue bool, value any)
	AppendCondition(cl types.Column, operation types.Operation, val any) error
}

type WhereBuilder struct {
	ctx    context.Context
	sql    string
	values []any
}

func NewWhereBuilder(ctx context.Context) *WhereBuilder {
	return &WhereBuilder{
		ctx: ctx,
	}
}

func (b *WhereBuilder) SQL() string {
	if strings.TrimSpace(b.sql) == "" {
		return ""
	}
	return " WHERE " + b.sql
}

func (b *WhereBuilder) Values() []any {
	return b.values
}

func (b *WhereBuilder) StartGroup() {
	if b.needANDBeforeCondition() {
		b.AND()
	}
	b.sql += "("
}
func (b *WhereBuilder) EndGroup() {
	b.sql += ")"
}

func (b *WhereBuilder) AND() {
	b.sql += " AND "
}

func (b *WhereBuilder) OR() {
	b.sql += " OR "
}

func (b *WhereBuilder) appendValue(val any) {
	switch values := val.(type) {
	case []any:
		if len(values) > 0 {
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

func (b *WhereBuilder) AppendSQLWithValues(sql string, appendValue bool, value any) {
	b.sql += sql
	if appendValue {
		b.appendValue(value)
	}
}

func (b *WhereBuilder) needANDBeforeCondition() bool {
	if len(b.sql) < 4 {
		return false
	}
	endsWith := b.sql[len(b.sql)-4:]
	switch endsWith {
	case "AND ", " OR ":
		return false
	default:
		// group starts 'AND' no needed
		if endsWith[len(endsWith)-1:] == "(" {
			return false
		}
	}
	return true
}

func (b *WhereBuilder) AppendCondition(cl types.Column, operation types.Operation, val any) error {
	filterFn, ok := cl.GetFilterFn(operation)
	if !ok {
		return fmt.Errorf("for field %s whereSQL %s option is not available", cl.GetField().GetStructPath(), operation)
	}
	sql, appendValue, err := filterFn(b.ctx, val)
	if err != nil {
		return err
	}
	if b.needANDBeforeCondition() {
		b.AND()
	}
	b.sql += sql
	if !appendValue {
		return nil
	}
	b.appendValue(val)
	return nil
}
