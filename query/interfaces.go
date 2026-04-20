package query

import "github.com/insei/gerpo/types"

// The interfaces in this file are small composable contracts that the
// per-operation helpers (GetFirstHelper, GetListHelper, …) embed. They are
// intentionally narrow so callers can write reusable middleware-style helpers
// without depending on the full operation interface.
//
// Example: a tenant-aware filter that works for every query type.
//
//	func applyTenant(h query.Filterable, tenantID uuid.UUID) {
//	    h.Where().Field(...).EQ(tenantID)
//	}
//
//	repo.GetFirst(ctx, func(m *User, h query.GetFirstHelper[User]) {
//	    applyTenant(h, tid)
//	})
//	repo.Update(ctx, &u, func(m *User, h query.UpdateHelper[User]) {
//	    applyTenant(h, tid)
//	})

// Filterable describes any helper that exposes a WHERE entry point.
// GetFirst, GetList, Count, Update and Delete all satisfy it.
type Filterable interface {
	// Where defines the starting point for building conditions in a query,
	// returning a types.WhereTarget interface.
	Where() types.WhereTarget
}

// Sortable describes any helper that exposes an ORDER BY entry point.
// GetFirst and GetList satisfy it.
type Sortable interface {
	// OrderBy defines the sorting criteria for a query and returns a
	// types.OrderTarget interface for further specification.
	OrderBy() types.OrderTarget
}

// Excludable describes any helper that lets the caller narrow the column set
// of an operation through Exclude / Only. GetFirst, GetList, Insert and
// Update satisfy it.
type Excludable interface {
	// Exclude removes specified fields from requesting data from repository storage.
	Exclude(fieldsPtr ...any)
	// Only includes the specified columns in the execution context, ignoring all others in the existing collection.
	Only(fieldsPtr ...any)
}

// Pageable describes the pagination contract on a list helper. The methods
// return GetListHelper so that the chain stays usable from inside a single
// query closure (h.Page(1).Size(20).Where()...).
type Pageable[TModel any] interface {
	// Page sets the page number for pagination in a query and returns the same GetListHelper instance.
	Page(page uint64) GetListHelper[TModel]
	// Size sets the maximum number of items to retrieve per page and returns the same GetListHelper instance.
	Size(size uint64) GetListHelper[TModel]
}

// Returnable describes any helper that exposes per-request control over the
// RETURNING clause. Insert and Update satisfy it.
//
// Default behavior: column-level markers (ReturnedOnInsert / ReturnedOnUpdate)
// decide what's returned. Calling Returning(...) overrides for this request:
// the listed fields become the returning set, replacing the defaults. Calling
// Returning() with no arguments disables RETURNING for the request.
type Returnable interface {
	// Returning narrows the RETURNING clause for this request. Calling with no arguments disables RETURNING;
	// calling with explicit fields replaces the repository's default returning set with exactly those columns.'
	Returning(fieldsPtr ...any)
}
