package gerpo

import (
	"context"

	"github.com/insei/gerpo/query"
)

type Option[TModel any] interface {
	apply(c *repository[TModel]) error
}

// optionFn is a type that implements the Option interface.
type optionFn[TModel any] func(c *repository[TModel]) error

// apply implements the Option interface for optionFn.
func (f optionFn[TModel]) apply(c *repository[TModel]) error { //nolint:unused // satisfies the Option interface used via dispatch
	return f(c)
}

// WithBeforeInsert registers a callback invoked right before a row is inserted.
// Returning a non-nil error from the callback aborts the Insert — the SQL is
// NOT executed and the error is returned to the caller (after passing through
// WithErrorTransformer, if any).
//
// Multiple registrations are chained in the order they were added; the first
// non-nil error stops the chain and becomes the returned value.
func WithBeforeInsert[TModel any](fn func(ctx context.Context, model *TModel) error) Option[TModel] {
	return optionFn[TModel](func(o *repository[TModel]) error {
		if fn == nil {
			return nil
		}
		if o.beforeInsert == nil {
			o.beforeInsert = fn
			return nil
		}
		wrap := o.beforeInsert
		o.beforeInsert = func(ctx context.Context, model *TModel) error {
			if err := wrap(ctx, model); err != nil {
				return err
			}
			return fn(ctx, model)
		}
		return nil
	})
}

// WithBeforeUpdate registers a callback invoked right before a row is updated.
// Returning a non-nil error aborts the Update — the SQL does NOT run.
// Chaining semantics match WithBeforeInsert.
func WithBeforeUpdate[TModel any](fn func(ctx context.Context, model *TModel) error) Option[TModel] {
	return optionFn[TModel](func(o *repository[TModel]) error {
		if fn == nil {
			return nil
		}
		if o.beforeUpdate == nil {
			o.beforeUpdate = fn
			return nil
		}
		wrap := o.beforeUpdate
		o.beforeUpdate = func(ctx context.Context, model *TModel) error {
			if err := wrap(ctx, model); err != nil {
				return err
			}
			return fn(ctx, model)
		}
		return nil
	})
}

// WithAfterSelect registers a callback invoked after GetFirst / GetList with
// the freshly scanned models. Returning a non-nil error surfaces it to the
// caller AFTER the rows have already been fetched — use this for cascade
// reads or context-level bookkeeping, not for validation.
func WithAfterSelect[TModel any](fn func(ctx context.Context, models []*TModel) error) Option[TModel] {
	return optionFn[TModel](func(o *repository[TModel]) error {
		if fn == nil {
			return nil
		}
		if o.afterSelect == nil {
			o.afterSelect = fn
			return nil
		}
		wrap := o.afterSelect
		o.afterSelect = func(ctx context.Context, models []*TModel) error {
			if err := wrap(ctx, models); err != nil {
				return err
			}
			return fn(ctx, models)
		}
		return nil
	})
}

// WithAfterInsert registers a callback invoked after a successful Insert.
// Returning a non-nil error surfaces it to the caller AFTER the row was
// already written. The row is NOT automatically rolled back — the caller
// decides whether to roll back an ambient transaction based on the error.
//
// Typical use: cascade related rows, emit an audit entry, publish an event
// in the same ctx-bound tx. See docs/features/hooks.md for the
// cascade-related-rows pattern.
func WithAfterInsert[TModel any](fn func(ctx context.Context, model *TModel) error) Option[TModel] {
	return optionFn[TModel](func(o *repository[TModel]) error {
		if fn == nil {
			return nil
		}
		if o.afterInsert == nil {
			o.afterInsert = fn
			return nil
		}
		wrap := o.afterInsert
		o.afterInsert = func(ctx context.Context, model *TModel) error {
			if err := wrap(ctx, model); err != nil {
				return err
			}
			return fn(ctx, model)
		}
		return nil
	})
}

// WithBeforeInsertMany registers a callback invoked right before a batch of
// rows is inserted via InsertMany. The callback receives the full slice in one
// call — typical use is bulk validation or resolving shared references in a
// single round-trip. Returning a non-nil error aborts the InsertMany; the SQL
// does NOT run.
//
// Chaining semantics match WithBeforeInsert.
func WithBeforeInsertMany[TModel any](fn func(ctx context.Context, models []*TModel) error) Option[TModel] {
	return optionFn[TModel](func(o *repository[TModel]) error {
		if fn == nil {
			return nil
		}
		if o.beforeInsertMany == nil {
			o.beforeInsertMany = fn
			return nil
		}
		wrap := o.beforeInsertMany
		o.beforeInsertMany = func(ctx context.Context, models []*TModel) error {
			if err := wrap(ctx, models); err != nil {
				return err
			}
			return fn(ctx, models)
		}
		return nil
	})
}

// WithAfterInsertMany registers a callback invoked after a successful
// InsertMany. The callback receives the full slice — use this for cascade
// inserts in one batched child query rather than calling the single-row hook
// once per parent.
//
// A non-nil error is surfaced AFTER the rows are already written; the caller
// decides whether to roll back an ambient transaction.
func WithAfterInsertMany[TModel any](fn func(ctx context.Context, models []*TModel) error) Option[TModel] {
	return optionFn[TModel](func(o *repository[TModel]) error {
		if fn == nil {
			return nil
		}
		if o.afterInsertMany == nil {
			o.afterInsertMany = fn
			return nil
		}
		wrap := o.afterInsertMany
		o.afterInsertMany = func(ctx context.Context, models []*TModel) error {
			if err := wrap(ctx, models); err != nil {
				return err
			}
			return fn(ctx, models)
		}
		return nil
	})
}

// WithAfterUpdate registers a callback invoked after a successful Update.
// Returning a non-nil error surfaces it to the caller AFTER the row was
// already modified — same contract as WithAfterInsert.
func WithAfterUpdate[TModel any](fn func(ctx context.Context, model *TModel) error) Option[TModel] {
	return optionFn[TModel](func(o *repository[TModel]) error {
		if fn == nil {
			return nil
		}
		if o.afterUpdate == nil {
			o.afterUpdate = fn
			return nil
		}
		wrap := o.afterUpdate
		o.afterUpdate = func(ctx context.Context, model *TModel) error {
			if err := wrap(ctx, model); err != nil {
				return err
			}
			return fn(ctx, model)
		}
		return nil
	})
}

// WithQuery applies a query function to configure query behavior in a repository instance.
func WithQuery[TModel any](queryFn func(m *TModel, h query.PersistentHelper[TModel])) Option[TModel] {
	return optionFn[TModel](func(r *repository[TModel]) error {
		if queryFn != nil {
			r.persistentQuery.HandleFn(queryFn)
		}
		return nil
	})
}

// WithErrorTransformer configures a repository to apply the provided error transformer function for error handling.
func WithErrorTransformer[TModel any](fn func(err error) error) Option[TModel] {
	return optionFn[TModel](func(o *repository[TModel]) error {
		if fn != nil {
			o.errorTransformer = fn
		}
		return nil
	})
}

// WithTracer installs a tracing hook called around every Repository operation.
// Pass nil to disable tracing (this is the default). The hook receives a
// SpanInfo with the operation name (prefixed with "gerpo.") and the bound
// table, and returns a context to propagate plus a SpanEnd that observes the
// terminal error.
func WithTracer[TModel any](tracer Tracer) Option[TModel] {
	return optionFn[TModel](func(o *repository[TModel]) error {
		o.tracer = tracer
		return nil
	})
}
