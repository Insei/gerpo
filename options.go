package gerpo

import (
	"context"
	"fmt"
	"strings"

	"github.com/insei/fmap/v3"
	"github.com/insei/gerpo/types"
)

type Option[TModel any] interface {
	apply(c *Repository[TModel])
}

// optionFn is a type that implements the Option interface.
type optionFn[TModel any] func(c *Repository[TModel])

// apply implements the Option interface for optionFn.
func (f optionFn[TModel]) apply(c *Repository[TModel]) {
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
	return optionFn[TModel](func(o *Repository[TModel]) {
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
	return optionFn[TModel](func(o *Repository[TModel]) {
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
	return optionFn[TModel](func(o *Repository[TModel]) {
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

//func WithSoftDelete[TModel any](fieldPtrFn func(d *TModel) any, valueFn func(ctx context.Context) any) Option[TModel] {
//	return optionFn[TModel](func(o *Repository[TModel]) {
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

func WithLeftJoin[TModel any](fn func(ctx context.Context) string) Option[TModel] {
	return optionFn[TModel](func(o *Repository[TModel]) {
		if fn != nil {
			if o.leftJoins == nil {
				o.leftJoins = fn
				return
			}
			wrap := o.leftJoins
			o.leftJoins = func(ctx context.Context) string {
				return strings.TrimSpace(fmt.Sprintf("%s %s", wrap(ctx), fn(ctx)))
			}
		}
	})
}

func WithTable[TModel any](table string) Option[TModel] {
	return optionFn[TModel](func(c *Repository[TModel]) {
		c.table = table
	})
}
