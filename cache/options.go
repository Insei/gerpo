package cache

type option func(c *sourceBundle)

// apply implements the Option interface for option.
// It calls the underlying function with the given sourceBundle.
func (f option) apply(c *sourceBundle) {
	f(c)
}

type Option interface {
	apply(c *sourceBundle)
}

func WithSource(s Source) Option {
	return option(func(m *sourceBundle) {
		if s != nil {
			m.sources = append(m.sources, s)
		}
	})
}
