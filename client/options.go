package bcln

type Options struct {
	UnexpectedResultCallback UnexpectedResultCallback
}

type SetOption func(o *Options)

// WithCallback sets the the unexpected result callback.
func WithUnexpectedResultCallback(callback UnexpectedResultCallback) SetOption {
	return func(o *Options) { o.UnexpectedResultCallback = callback }
}

func Apply(ops []SetOption, c *Options) {
	for i := range ops {
		if ops[i] != nil {
			ops[i](c)
		}
	}
}
