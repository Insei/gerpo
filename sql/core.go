package sql

import (
	"context"
	"fmt"
	"strings"
)

type StringBuilder struct {
	ctx           context.Context
	joinsSQL      string
	table         string
	orderBuilder  *StringOrderBuilder
	whereBuilder  *StringWhereBuilder
	selectBuilder *StringSelectBuilder
	groupBuilder  *StringGroupBuilder
	insertBuilder *StringInsertBuilder
	updateBuilder *StringUpdateBuilder
}

func (b *StringBuilder) WhereBuilder() *StringWhereBuilder {
	return b.whereBuilder
}

func (b *StringBuilder) OrderBuilder() *StringOrderBuilder {
	return b.orderBuilder
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

func (b *StringBuilder) selectSQL(selectSQL, whereSQL, orderSQL, groupSQL, joinSQL,
	limitNumStr, offsetNumStr string) (string, []any) {
	sql := fmt.Sprintf("SELECT %s FROM %s", selectSQL, b.table)
	if strings.TrimSpace(whereSQL) != "" {
		sql += fmt.Sprintf(" WHERE %s", whereSQL)
	}

	if strings.TrimSpace(orderSQL) != "" {
		sql += fmt.Sprintf(" ORDER BY %s", orderSQL)
	}
	if strings.TrimSpace(groupSQL) != "" {
		sql += fmt.Sprintf(" GROUP BY %s", groupSQL)
	}
	if strings.TrimSpace(joinSQL) != "" {
		sql += fmt.Sprintf(" %s", joinSQL)
	}
	if strings.TrimSpace(limitNumStr) != "" {
		sql += fmt.Sprintf(" LIMIT %s", limitNumStr)
	}
	if strings.TrimSpace(offsetNumStr) != "" {
		sql += fmt.Sprintf(" OFFSET %s", offsetNumStr)
	}
	return sql, b.whereBuilder.values
}

func (b *StringBuilder) CountSQL() (string, []any) {
	b.selectBuilder.Limit(1)
	return b.selectSQL("count(*) over() AS count", b.whereBuilder.sql,
		b.orderBuilder.sql, b.groupBuilder.sql, b.joinsSQL, b.selectBuilder.GetLimit(), b.selectBuilder.GetOffset())
}

func (b *StringBuilder) SelectSQL() (string, []any) {
	return b.selectSQL(b.selectBuilder.GetSQL(), b.whereBuilder.sql, b.orderBuilder.sql, b.groupBuilder.sql, b.joinsSQL,
		b.selectBuilder.GetLimit(), b.selectBuilder.GetOffset())
}

func (b *StringBuilder) InsertSQL() string {
	return fmt.Sprintf("INSERT INTO %s %s", b.table, b.insertBuilder.SQL())
}

func (b *StringBuilder) UpdateSQL() string {
	sql := fmt.Sprintf("UPDATE %s SET %s", b.table, b.updateBuilder.SQL())
	if b.whereBuilder.sql != "" {
		sql += fmt.Sprintf(" WHERE %s", b.whereBuilder.sql)
	}
	return sql
}

func (b *StringBuilder) DeleteSQL() string {
	sql := fmt.Sprintf("DELETE FROM %s", b.table)
	if b.whereBuilder.sql == "" {
		panic(fmt.Errorf("delete all table rows in not available"))
	}
	sql += fmt.Sprintf(" WHERE %s", b.whereBuilder.sql)
	return sql
}

func NewStringBuilder(ctx context.Context, table string) *StringBuilder {
	return &StringBuilder{
		table: table,
		ctx:   ctx,
		whereBuilder: &StringWhereBuilder{
			ctx: ctx,
		},
		orderBuilder: &StringOrderBuilder{
			ctx: ctx,
		},
		selectBuilder: &StringSelectBuilder{
			ctx: ctx,
		},
		groupBuilder: &StringGroupBuilder{
			ctx: ctx,
		},
		insertBuilder: &StringInsertBuilder{
			ctx: ctx,
		},
		updateBuilder: &StringUpdateBuilder{
			ctx: ctx,
		},
	}
}
