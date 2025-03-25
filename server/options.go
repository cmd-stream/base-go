package bser

import "time"

type Options struct {
	WorkersCount     int
	LostConnCallback LostConnCallback
	ConnReceiver     []SetConnReceiverOption
}

type SetOption func(o *Options)

// WithWorkersCount sets the number of workers. Must be greater than 0.
func WithWorkersCount(count int) SetOption {
	return func(o *Options) { o.WorkersCount = count }
}

// WithLostConnCallback sets the callback function to be invoked
// when a connection is lost.
func WithLostConnCallback(callback LostConnCallback) SetOption {
	return func(o *Options) { o.LostConnCallback = callback }
}

// WithConnReceiver configures the ConnReceiver with the specified options.
func WithConnReceiver(ops ...SetConnReceiverOption) SetOption {
	return func(o *Options) { o.ConnReceiver = ops }
}

func Apply(ops []SetOption, o *Options) {
	for i := range ops {
		if ops[i] != nil {
			ops[i](o)
		}
	}
}

type ConnReceiverOptions struct {
	FirstConnTimeout time.Duration
}

type SetConnReceiverOption func(o *ConnReceiverOptions)

// WithFirstConnTimeout sets the timeout for the first connection attempt.
func WithFirstConnTimeout(d time.Duration) SetConnReceiverOption {
	return func(o *ConnReceiverOptions) { o.FirstConnTimeout = d }
}

func ApplyForConnReceiver(ops []SetConnReceiverOption, o *ConnReceiverOptions) {
	for i := range ops {
		if ops[i] != nil {
			ops[i](o)
		}
	}
}
