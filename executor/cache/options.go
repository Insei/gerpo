package cache

type option func(c *storagesBundle)

// apply implements the Option interface for option.
// It calls the underlying function with the given storagesBundle.
func (f option) apply(c *storagesBundle) {
	f(c)
}

type Option interface {
	apply(c *storagesBundle)
}

func WithStorage(s Storage) Option {
	return option(func(m *storagesBundle) {
		if s != nil {
			m.storages = append(m.storages, s)
		}
	})
}
