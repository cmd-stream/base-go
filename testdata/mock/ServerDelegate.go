package mock

import (
	"context"
	"net"

	"github.com/ymz-ncnk/mok"
)

func NewServerDelegate() ServerDelegate {
	return ServerDelegate{
		Mock: mok.New("ServerDelegate"),
	}
}

type ServerDelegate struct {
	*mok.Mock
}

func (m ServerDelegate) RegisterHandle(
	fn func(ctx context.Context, conn net.Conn) (err error)) ServerDelegate {
	m.Register("Handle", fn)
	return m
}

func (m ServerDelegate) RegisterNHandle(n int,
	fn func(ctx context.Context, conn net.Conn) (err error)) ServerDelegate {
	m.RegisterN("Handle", n, fn)
	return m
}

func (m ServerDelegate) Handle(ctx context.Context, conn net.Conn) (err error) {
	vals, err := m.Call("Handle", mok.SafeVal[context.Context](ctx),
		mok.SafeVal[net.Conn](conn))
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}
