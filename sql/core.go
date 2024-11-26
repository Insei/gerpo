package sql

import (
	"context"
	"fmt"
	"strings"

	"github.com/insei/gerpo/types"
)

type StringBuilder struct {
	ctx           context.Context
	table         string
	whereBuilder  *StringWhereBuilder
	selectBuilder *StringSelectBuilder
	groupBuilder  *StringGroupBuilder
	insertBuilder *StringInsertBuilder
	updateBuilder *StringUpdateBuilder
	joinBuilder   *StringJoinBuilder
}

func (b *StringBuilder) WhereBuilder() *StringWhereBuilder {
	return b.whereBuilder
}

func (b *StringBuilder) GroupBuilder() *StringGroupBuilder {
	return b.groupBuilder
}

func (b *StringBuilder) SelectBuilder() *StringSelectBuilder {
	return b.selectBuilder
}

func (b *StringBuilder) InsertBuilder() *StringInsertBuilder {
	return b.insertBuilder
}

func (b *StringBuilder) UpdateBuilder() *StringUpdateBuilder {
	return b.updateBuilder
}

func (b *StringBuilder) JoinBuilder() *StringJoinBuilder {
	return b.joinBuilder
}

func (b *StringBuilder) selectSQLBase(selectedColumns string) string {
	sql := fmt.Sprintf("SELECT %s FROM %s", selectedColumns, b.table)
	joinSQL := b.joinBuilder.SQL()
	if strings.TrimSpace(joinSQL) != "" {
		sql += fmt.Sprintf(" %s", joinSQL)
	}
	whereSQL := b.whereBuilder.SQL()
	if strings.TrimSpace(whereSQL) != "" {
		sql += fmt.Sprintf(" WHERE %s", whereSQL)
	}
	orderSQL := b.selectBuilder.GetOrderSQL()
	if strings.TrimSpace(orderSQL) != "" {
		sql += fmt.Sprintf(" ORDER BY %s", orderSQL)
	}
	groupSQL := b.groupBuilder.SQL()
	if strings.TrimSpace(groupSQL) != "" {
		sql += fmt.Sprintf(" GROUP BY %s", groupSQL)
	}
	limitNumStr := b.selectBuilder.GetLimit()
	if strings.TrimSpace(limitNumStr) != "" {
		sql += fmt.Sprintf(" LIMIT %s", limitNumStr)
	}
	offsetNumStr := b.selectBuilder.GetOffset()
	if strings.TrimSpace(offsetNumStr) != "" {
		sql += fmt.Sprintf(" OFFSET %s", offsetNumStr)
	}
	return sql
}

func (b *StringBuilder) countSQL() string {
	b.selectBuilder.Limit(1)
	return b.selectSQLBase("count(*) over() AS count")
}

func (b *StringBuilder) selectSQL() string {
	return b.selectSQLBase(b.selectBuilder.GetSQL())
}

func (b *StringBuilder) insertSQL() string {
	return fmt.Sprintf("INSERT INTO %s %s", b.table, b.insertBuilder.SQL())
}

func (b *StringBuilder) updateSQL() string {
	sql := fmt.Sprintf("UPDATE %s SET %s", b.table, b.updateBuilder.SQL())
	if b.whereBuilder.sql != "" {
		sql += fmt.Sprintf(" WHERE %s", b.whereBuilder.sql)
	}
	return sql
}

func (b *StringBuilder) deleteSQL() string {
	sql := fmt.Sprintf("DELETE FROM %s", b.table)
	joinSQL := b.joinBuilder.SQL()
	if strings.TrimSpace(joinSQL) != "" {
		sql += fmt.Sprintf(" %s", joinSQL)
	}
	if b.whereBuilder.sql == "" {
		panic(fmt.Errorf("delete all table rows in not available"))
	}
	sql += fmt.Sprintf(" WHERE %s", b.whereBuilder.sql)
	return sql
}

func (b *StringBuilder) GetStmtWithArgs(operation Operation) (string, []any) {
	switch operation {
	case Select:
		return b.selectSQL(), b.WhereBuilder().Values()
	case SelectOne:
		b.selectBuilder.Limit(1)
		return b.selectSQL(), b.WhereBuilder().Values()
	case Count:
		return b.countSQL(), b.WhereBuilder().Values()
	case Delete:
		return b.deleteSQL(), b.WhereBuilder().Values()
	case Insert:
		panic(fmt.Errorf("insert operation is not available"))
	case Update:
		panic(fmt.Errorf("update operation is not available"))
	default:
		panic(fmt.Errorf("unrecognized operation"))
	}
}

func (b *StringBuilder) GetStmtWithArgsForModel(operation Operation, model any) (string, []any) {
	switch operation {
	case Insert:
		values := b.InsertBuilder().GetColumnValues(model)
		return b.insertSQL(), values
	case Update:
		values := b.UpdateBuilder().GetColumnValues(model)
		whereValues := b.WhereBuilder().Values()
		return b.updateSQL(), append(values, whereValues...)
	default:
		panic(fmt.Errorf("unrecognized or unsuported operation"))
	}
}

func (b *StringBuilder) GetModelPointers(operation Operation, model any) []any {
	switch operation {
	case Select, SelectOne:
		return b.selectBuilder.GetColumnFieldPointers(model)
	default:
		panic(fmt.Errorf("unrecognized or unsuported operation"))
	}
}

func NewStringBuilder(ctx context.Context, table string, columns *types.ColumnsStorage) *StringBuilder {
	return &StringBuilder{
		table: table,
		ctx:   ctx,
		whereBuilder: &StringWhereBuilder{
			ctx: ctx,
		},
		selectBuilder: &StringSelectBuilder{
			ctx:     ctx,
			columns: columns.AsSliceByAction(types.SQLActionSelect),
		},
		groupBuilder: &StringGroupBuilder{
			ctx: ctx,
		},
		insertBuilder: &StringInsertBuilder{
			ctx:     ctx,
			columns: columns.AsSliceByAction(types.SQLActionInsert),
		},
		updateBuilder: &StringUpdateBuilder{
			ctx:     ctx,
			columns: columns.AsSliceByAction(types.SQLActionUpdate),
		},
		joinBuilder: &StringJoinBuilder{
			ctx: ctx,
		},
	}
}

type StringBuilderFactory func(ctx context.Context) *StringBuilder

func (b StringBuilderFactory) New(ctx context.Context) *StringBuilder {
	return b(ctx)
}

func NewStringBuilderFactory(table string, columns *types.ColumnsStorage) StringBuilderFactory {
	return func(ctx context.Context) *StringBuilder {
		return NewStringBuilder(ctx, table, columns)
	}
}
