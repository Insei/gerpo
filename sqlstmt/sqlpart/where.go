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
			return query + " IS NULL", false
		}
		return query + " = ?", true
	}
}
func genNotEQFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		if value == nil {
			return query + " IS NOT NULL", false
		}
		return query + " != ?", true
	}
}

func genLTFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return query + " < ?", true
	}
}
func genLTEFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return query + " <= ?", true
	}
}

func genGTFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return query + " > ?", true
	}
}
func genGTEFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return query + " >= ?", true
	}
}

func genInFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		fPtr := ((*[2]unsafe.Pointer)(unsafe.Pointer(&value)))[1]
		anyArr := (*[]any)(fPtr)
		if value == nil || len(*anyArr) == 0 {
			return "1 = 2", false
		}
		placeholders := strings.Repeat("?,", len(*anyArr))
		placeholders = placeholders[:len(placeholders)-1]
		return query + " IN (" + placeholders + ")", true
	}
}
func genNotInFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		fPtr := ((*[2]unsafe.Pointer)(unsafe.Pointer(&value)))[1]
		anyArr := (*[]any)(fPtr)
		if value == nil || len(*anyArr) == 0 {
			return "1 = 1", false
		}
		placeholders := strings.Repeat("?,", len(*anyArr))
		placeholders = placeholders[:len(placeholders)-1]
		return query + " NOT IN (" + placeholders + ")", true
	}
}

// В LIKE-операторах параметр обёрнут в CAST(? AS text), чтобы PostgreSQL мог
// вывести тип параметра в CONCAT-контексте. CAST(? AS text) работает одинаково
// в PostgreSQL и MySQL, поэтому переносимость сохраняется.

func genContainsFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return query + " LIKE CONCAT('%', CAST(? AS text), '%')", true
	}
}
func genNotContainsFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return query + " NOT LIKE CONCAT('%', CAST(? AS text), '%')", true
	}
}

func genStartsWithFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return query + " LIKE CONCAT(CAST(? AS text), '%')", true
	}
}

func genNotStartsWithFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return query + " NOT LIKE CONCAT(CAST(? AS text), '%')", true
	}
}

func genEndsWithFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return query + " LIKE CONCAT('%', CAST(? AS text))", true
	}
}

func genNotEndsWithFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return query + " NOT LIKE CONCAT('%', CAST(? AS text))", true
	}
}

// Case-insensitive "fold" variants — mirrors strings.EqualFold naming.

func genEQFoldFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		if value == nil {
			return query + " IS NULL", false
		}
		return "LOWER(" + query + ") = LOWER(CAST(? AS text))", true
	}
}
func genNotEQFoldFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		if value == nil {
			return query + " IS NOT NULL", false
		}
		return "LOWER(" + query + ") != LOWER(CAST(? AS text))", true
	}
}

func genContainsFoldFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return "LOWER(" + query + ") LIKE LOWER(CONCAT('%', CAST(? AS text), '%'))", true
	}
}
func genNotContainsFoldFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return "LOWER(" + query + ") NOT LIKE LOWER(CONCAT('%', CAST(? AS text), '%'))", true
	}
}

func genStartsWithFoldFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return "LOWER(" + query + ") LIKE LOWER(CONCAT(CAST(? AS text), '%'))", true
	}
}
func genNotStartsWithFoldFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return "LOWER(" + query + ") NOT LIKE LOWER(CONCAT(CAST(? AS text), '%'))", true
	}
}

func genEndsWithFoldFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return "LOWER(" + query + ") LIKE LOWER(CONCAT('%', CAST(? AS text)))", true
	}
}
func genNotEndsWithFoldFn(query string) func(ctx context.Context, value any) (string, bool) {
	return func(ctx context.Context, value any) (string, bool) {
		return "LOWER(" + query + ") NOT LIKE LOWER(CONCAT('%', CAST(? AS text)))", true
	}
}

func GetFieldTypeFilters(field fmap.Field, sqlColumnString string) map[types.Operation]func(ctx context.Context, value any) (string, bool) {
	filters := make(map[types.Operation]func(ctx context.Context, value any) (string, bool))
	if field.GetType().Kind() == reflect.Ptr {
		filters[types.OperationEQ] = genEQFn(sqlColumnString)
		filters[types.OperationNotEQ] = genNotEQFn(sqlColumnString)
	}

	derefType := field.GetDereferencedType()
	switch derefType.Kind() {
	case reflect.Bool:
		filters[types.OperationEQ] = genEQFn(sqlColumnString)
		filters[types.OperationNotEQ] = genNotEQFn(sqlColumnString)
	case reflect.String:
		filters[types.OperationEQ] = genEQFn(sqlColumnString)
		filters[types.OperationNotEQ] = genNotEQFn(sqlColumnString)
		filters[types.OperationIn] = genInFn(sqlColumnString)
		filters[types.OperationNotIn] = genNotInFn(sqlColumnString)
		filters[types.OperationContains] = genContainsFn(sqlColumnString)
		filters[types.OperationNotContains] = genNotContainsFn(sqlColumnString)
		filters[types.OperationStartsWith] = genStartsWithFn(sqlColumnString)
		filters[types.OperationNotStartsWith] = genNotStartsWithFn(sqlColumnString)
		filters[types.OperationEndsWith] = genEndsWithFn(sqlColumnString)
		filters[types.OperationNotEndsWith] = genNotEndsWithFn(sqlColumnString)
		filters[types.OperationEQFold] = genEQFoldFn(sqlColumnString)
		filters[types.OperationNotEQFold] = genNotEQFoldFn(sqlColumnString)
		filters[types.OperationContainsFold] = genContainsFoldFn(sqlColumnString)
		filters[types.OperationNotContainsFold] = genNotContainsFoldFn(sqlColumnString)
		filters[types.OperationStartsWithFold] = genStartsWithFoldFn(sqlColumnString)
		filters[types.OperationNotStartsWithFold] = genNotStartsWithFoldFn(sqlColumnString)
		filters[types.OperationEndsWithFold] = genEndsWithFoldFn(sqlColumnString)
		filters[types.OperationNotEndsWithFold] = genNotEndsWithFoldFn(sqlColumnString)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		filters[types.OperationEQ] = genEQFn(sqlColumnString)
		filters[types.OperationNotEQ] = genNotEQFn(sqlColumnString)
		filters[types.OperationLT] = genLTFn(sqlColumnString)
		filters[types.OperationLTE] = genLTEFn(sqlColumnString)
		filters[types.OperationGT] = genGTFn(sqlColumnString)
		filters[types.OperationGTE] = genGTEFn(sqlColumnString)
		filters[types.OperationIn] = genInFn(sqlColumnString)
		filters[types.OperationNotIn] = genNotInFn(sqlColumnString)
	default:
		switch derefType {
		case reflect.TypeOf(time.Time{}):
			filters[types.OperationLT] = genLTFn(sqlColumnString)
			filters[types.OperationGT] = genGTFn(sqlColumnString)
			filters[types.OperationLTE] = genLTEFn(sqlColumnString)
			filters[types.OperationGTE] = genGTEFn(sqlColumnString)
		case reflect.TypeOf(uuid.UUID{}):
			filters[types.OperationEQ] = genEQFn(sqlColumnString)
			filters[types.OperationNotEQ] = genNotEQFn(sqlColumnString)
			filters[types.OperationIn] = genInFn(sqlColumnString)
			filters[types.OperationNotIn] = genNotInFn(sqlColumnString)
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
	sql    []byte
	values []any
}

func NewWhereBuilder(ctx context.Context) *WhereBuilder {
	return &WhereBuilder{
		ctx: ctx,
	}
}

// Reset prepares the builder for reuse by a new query without dropping underlying buffers.
func (b *WhereBuilder) Reset(ctx context.Context) {
	b.ctx = ctx
	b.sql = b.sql[:0]
	b.values = b.values[:0]
}

func (b *WhereBuilder) SQL() string {
	if len(b.sql) < 1 {
		return ""
	}
	return " WHERE " + string(b.sql)
}

func (b *WhereBuilder) Values() []any {
	return b.values
}

func (b *WhereBuilder) StartGroup() {
	if b.needANDBeforeCondition() {
		b.AND()
	}
	b.sql = append(b.sql, '(')
}
func (b *WhereBuilder) EndGroup() {
	b.sql = append(b.sql, ')')
}

func (b *WhereBuilder) AND() {
	b.sql = append(b.sql, " AND "...)
}

func (b *WhereBuilder) OR() {
	b.sql = append(b.sql, " OR "...)
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
	b.sql = append(b.sql, sql...)
	if appendValue {
		b.appendValue(value)
	}
}

func (b *WhereBuilder) needANDBeforeCondition() bool {
	if len(b.sql) < 4 {
		return false
	}
	endsWith := string(b.sql[len(b.sql)-4:])
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

// columnSQLArgsProvider mirrors sqlstmt.columnArgsProvider so the where-builder
// can prepend Compute-bound args without taking a dependency on the higher-level
// sqlstmt package.
type columnSQLArgsProvider interface {
	SQLArgs() []any
}

func (b *WhereBuilder) AppendCondition(cl types.Column, operation types.Operation, val any) error {
	if cl.IsAggregate() && !cl.HasFilterOverride(operation) {
		return fmt.Errorf("aggregate virtual column %q cannot be filtered without an explicit Filter() override (op=%s)",
			cl.GetField().GetStructPath(), operation)
	}
	filterFn, ok := cl.GetFilterFn(operation)
	if !ok {
		return fmt.Errorf("for field %s whereSQL %s option is not available", cl.GetField().GetStructPath(), operation)
	}
	sql, args, err := filterFn(b.ctx, val)
	if err != nil {
		return err
	}
	if sql == "" {
		return nil
	}
	if b.needANDBeforeCondition() {
		b.AND()
	}
	b.sql = append(b.sql, sql...)
	// Auto-derived filters wrap the column expression as `(compute_sql) op ?`, so any
	// bound args belonging to compute_sql must appear *before* the user value. Custom
	// filter overrides own their SQL entirely and decide whether to include those args.
	if !cl.HasFilterOverride(operation) {
		if ap, ok := cl.(columnSQLArgsProvider); ok {
			for _, a := range ap.SQLArgs() {
				b.appendValue(a)
			}
		}
	}
	for _, a := range args {
		b.appendValue(a)
	}
	return nil
}
