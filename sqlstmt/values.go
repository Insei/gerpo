package sqlstmt

import "github.com/insei/gerpo/types"

type values struct {
	values []any

	columns types.ExecutionColumns
}

func newValues(columns types.ExecutionColumns) *values {
	return &values{
		columns: columns,
	}
}

type Option func(v *values)

func (o Option) Apply(v *values) {
	o(v)
}

func WithModelValues(model any) Option {
	return func(v *values) {
		vals := v.columns.GetModelValues(model)
		v.values = append(vals, v.values...)
	}
}
