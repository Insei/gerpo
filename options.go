package gerpo

import (
	"context"

	"github.com/insei/gerpo/query"
)

type Option[TModel any] interface {
	apply(c *repository[TModel])
}

// optionFn is a type that implements the Option interface.
type optionFn[TModel any] func(c *repository[TModel])

// apply implements the Option interface for optionFn.
func (f optionFn[TModel]) apply(c *repository[TModel]) {
	f(c)
}

func WithBeforeInsert[TModel any](fn func(ctx context.Context, model *TModel)) Option[TModel] {
	return optionFn[TModel](func(o *repository[TModel]) {
		if fn != nil {
			if o.beforeInsert == nil {
				o.beforeInsert = fn
				return
			}
			wrap := o.beforeInsert
			o.beforeInsert = func(ctx context.Context, model *TModel) {
				wrap(ctx, model)
				fn(ctx, model)
			}
		}
	})
}

func WithBeforeUpdate[TModel any](fn func(ctx context.Context, model *TModel)) Option[TModel] {
	return optionFn[TModel](func(o *repository[TModel]) {
		if fn != nil {
			if o.beforeUpdate == nil {
				o.beforeUpdate = fn
				return
			}
			wrap := o.beforeUpdate
			o.beforeUpdate = func(ctx context.Context, model *TModel) {
				wrap(ctx, model)
				fn(ctx, model)
			}
		}
	})
}

func WithAfterSelect[TModel any](fn func(ctx context.Context, models []*TModel)) Option[TModel] {
	return optionFn[TModel](func(o *repository[TModel]) {
		if fn != nil {
			if o.afterSelect == nil {
				o.afterSelect = fn
				return
			}
			wrap := o.afterSelect
			o.afterSelect = func(ctx context.Context, model []*TModel) {
				wrap(ctx, model)
				fn(ctx, model)
			}
		}
	})
}

func WithQuery[TModel any](queryFn func(m *TModel, h query.PersistentUserHelper[TModel])) Option[TModel] {
	return optionFn[TModel](func(o *repository[TModel]) {
		if queryFn != nil {
			o.query.Persistent(queryFn)
		}
	})
}
