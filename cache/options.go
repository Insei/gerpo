package cache

type option func(c *modelBundle)

// apply implements the Option interface for option.
// It calls the underlying function with the given Column.
func (f option) apply(c *modelBundle) {
	f(c)
}

type Option interface {
	apply(c *modelBundle)
}

func WithSource(s Source) Option {
	return option(func(m *modelBundle) {
		if s != nil {
			m.sources = append(m.sources, s)
		}
	})
}
