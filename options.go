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

// WithBeforeInsert sets a function to be executed before inserting a model into the repository.
// If an existing function is already set, the new function will wrap the existing one, executing both in sequence.
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

// WithBeforeUpdate registers a function to be invoked before the update operation on the specified model in the repository.
// If an existing function is already set, the new function will wrap the existing one, executing both in sequence.
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

// WithAfterSelect returns an Option that appends or assigns a callback executed after select queries in the repository.
// If an existing function is already set, the new function will wrap the existing one, executing both in sequence.
func WithAfterSelect[TModel any](fn func(ctx context.Context, models []*TModel)) Option[TModel] {
	return optionFn[TModel](func(o *repository[TModel]) {
		if fn != nil {
			if o.afterSelect == nil {
				o.afterSelect = fn
				return
			}
			wrap := o.afterSelect
			o.afterSelect = func(ctx context.Context, models []*TModel) {
				wrap(ctx, models)
				fn(ctx, models)
			}
		}
	})
}

// WithAfterInsert creates an option to set a callback function that is executed after an insert operation in the repository.
// If an existing function is already set, the new function will wrap the existing one, executing both in sequence.
func WithAfterInsert[TModel any](fn func(ctx context.Context, model *TModel)) Option[TModel] {
	return optionFn[TModel](func(o *repository[TModel]) {
		if fn != nil {
			if o.afterInsert == nil {
				o.afterInsert = fn
				return
			}
			wrap := o.afterInsert
			o.afterInsert = func(ctx context.Context, model *TModel) {
				wrap(ctx, model)
				fn(ctx, model)
			}
		}
	})
}

// WithAfterUpdate creates an Option to set or append a callback function that triggers after an update operation on the model.
// If an existing function is already set, the new function will wrap the existing one, executing both in sequence.
func WithAfterUpdate[TModel any](fn func(ctx context.Context, model *TModel)) Option[TModel] {
	return optionFn[TModel](func(o *repository[TModel]) {
		if fn != nil {
			if o.afterUpdate == nil {
				o.afterUpdate = fn
				return
			}
			wrap := o.afterUpdate
			o.afterUpdate = func(ctx context.Context, model *TModel) {
				wrap(ctx, model)
				fn(ctx, model)
			}
		}
	})
}

// WithQuery applies a query function to configure query behavior in a repository instance.
func WithQuery[TModel any](queryFn func(m *TModel, h query.PersistentHelper[TModel])) Option[TModel] {
	return optionFn[TModel](func(r *repository[TModel]) {
		if queryFn != nil {
			r.persistentQuery.HandleFn(queryFn)
		}
	})
}

// WithErrorTransformer configures a repository to apply the provided error transformer function for error handling.
func WithErrorTransformer[TModel any](fn func(err error) error) Option[TModel] {
	return optionFn[TModel](func(o *repository[TModel]) {
		if fn != nil {
			o.errorTransformer = fn
		}
	})
}
