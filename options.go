package gerpo

import (
	"context"

	"github.com/insei/fmap/v3"
	"github.com/insei/gerpo/query"
	"github.com/insei/gerpo/types"
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

type options[TModel any] struct {
	model        *TModel
	fields       fmap.Storage
	columns      *types.ColumnsStorage
	beforeInsert []func(ctx context.Context, model *TModel)
	beforeUpdate []func(ctx context.Context, model *TModel)
	afterSelect  []func(ctx context.Context, models []*TModel)
	leftJoins    []func(ctx context.Context) string
	softDelete   map[types.Column]func(ctx context.Context) any
}

func newOptions[TModel any](model *TModel, columns *types.ColumnsStorage, fields fmap.Storage) *options[TModel] {
	return &options[TModel]{
		model:        model,
		fields:       fields,
		columns:      columns,
		beforeInsert: nil,
		beforeUpdate: nil,
		afterSelect:  nil,
		leftJoins:    nil,
		softDelete:   make(map[types.Column]func(ctx context.Context) any),
	}
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

//func WithSoftDelete[TModel any](fieldPtrFn func(d *TModel) any, valueFn func(ctx context.Context) any) Option[TModel] {
//	return optionFn[TModel](func(o *repository[TModel]) {
//		field, err := o.fields.GetFieldByPtr(o.model, fieldPtrFn(o.model))
//		if err != nil {
//			panic(err)
//		}
//		cl, ok := o.columns.Get(field)
//		if !ok {
//			panic("cannot find column for soft deletion setup")
//		}
//		if !cl.IsAllowedAction(types.SQLActionUpdate) {
//			panic(fmt.Errorf("cannot setup soft deletion with %s field, update is not supported", field.GetStructPath()))
//		}
//		o.softDelete[cl] = valueFn
//	})
//}
