package mock

import (
	"context"
	"net"
	"reflect"

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
	var ctxVal reflect.Value
	if ctx == nil {
		ctxVal = reflect.Zero(reflect.TypeOf((*context.Context)(nil)).Elem())
	} else {
		ctxVal = reflect.ValueOf(ctx)
	}
	var connVal reflect.Value
	if conn == nil {
		connVal = reflect.Zero(reflect.TypeOf((*net.Conn)(nil)).Elem())
	} else {
		connVal = reflect.ValueOf(conn)
	}
	vals, err := m.Call("Handle", ctxVal, connVal)
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}
