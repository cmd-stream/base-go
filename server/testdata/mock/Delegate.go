package mock

import (
	"context"
	"net"

	"github.com/ymz-ncnk/mok"
)

type HandleFn func(ctx context.Context, conn net.Conn) (err error)

func NewDelegate() Delegate {
	return Delegate{Mock: mok.New("Delegate")}
}

type Delegate struct {
	*mok.Mock
}

func (m Delegate) RegisterHandle(fn HandleFn) Delegate {
	m.Register("Handle", fn)
	return m
}

func (m Delegate) RegisterNHandle(n int, fn HandleFn) Delegate {
	m.RegisterN("Handle", n, fn)
	return m
}

func (m Delegate) Handle(ctx context.Context, conn net.Conn) (err error) {
	vals, err := m.Call("Handle", mok.SafeVal[context.Context](ctx),
		mok.SafeVal[net.Conn](conn))
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}
