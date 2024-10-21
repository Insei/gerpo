package api

type FieldLinkConnectorFn func(dtoFieldPtr, modelFieldPtr any) error

func (f FieldLinkConnectorFn) Link(dtoFieldPtr, modelFieldPtr any) error {
	return f(dtoFieldPtr, modelFieldPtr)
}

type advancedLinksOption[TDto, TModel any] struct {
	fn func(dto *TDto, model *TModel, conn FieldConnector)
}

func (o *advancedLinksOption[TDto, TModel]) apply(_ *core) {

}

type APIConnectorOption[TModel any] interface {
	apply(c *core)
}

func WithAdvancedFieldLink[TDto, TModel any](fn func(dto *TDto, model *TModel, conn FieldConnector)) APIConnectorOption[TModel] {
	return &advancedLinksOption[TDto, TModel]{fn: fn}
}
