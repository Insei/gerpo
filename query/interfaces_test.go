package query

// Compile-time guarantees that concrete helpers implement the small composable
// contracts as well as their per-operation aggregates. If any of these break,
// the corresponding concrete type drifted from the interface contract.
var (
	_ Filterable          = (*GetFirst[any])(nil)
	_ Sortable            = (*GetFirst[any])(nil)
	_ Excludable          = (*GetFirst[any])(nil)
	_ GetFirstHelper[any] = (*GetFirst[any])(nil)

	_ Filterable         = (*GetList[any])(nil)
	_ Sortable           = (*GetList[any])(nil)
	_ Excludable         = (*GetList[any])(nil)
	_ Pageable[any]      = (*GetList[any])(nil)
	_ GetListHelper[any] = (*GetList[any])(nil)

	_ Filterable       = (*Count[any])(nil)
	_ CountHelper[any] = (*Count[any])(nil)

	_ Excludable        = (*Insert[any])(nil)
	_ InsertHelper[any] = (*Insert[any])(nil)

	_ Filterable        = (*Update[any])(nil)
	_ Excludable        = (*Update[any])(nil)
	_ UpdateHelper[any] = (*Update[any])(nil)

	_ Filterable        = (*Delete[any])(nil)
	_ DeleteHelper[any] = (*Delete[any])(nil)
)
