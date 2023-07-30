package mock

import (
	"context"
	"reflect"
	"time"

	"github.com/cmd-stream/base-go"
	"github.com/ymz-ncnk/mok"
)

func NewCmd() Cmd {
	return Cmd{
		Mock: mok.New("Cmd"),
	}
}

type Cmd struct {
	*mok.Mock
}

func (c Cmd) RegisterExec(
	fn func(ctx context.Context, at time.Time, seq base.Seq, receiver any,
		proxy base.Proxy) (err error)) Cmd {
	c.Register("Exec", fn)
	return c
}

func (c Cmd) RegisterTimeout(
	fn func() (timeout time.Duration)) Cmd {
	c.Register("Timeout", fn)
	return c
}

func (c Cmd) Exec(ctx context.Context, at time.Time, seq base.Seq, receiver any,
	proxy base.Proxy) (err error) {
	var ctxVal reflect.Value
	if ctx == nil {
		ctxVal = reflect.Zero(reflect.TypeOf((*context.Context)(nil)).Elem())
	} else {
		ctxVal = reflect.ValueOf(ctx)
	}
	var receiverVal any
	if receiver == nil {
		receiverVal = reflect.Zero(reflect.TypeOf((*any)(nil)).Elem())
	} else {
		receiverVal = reflect.ValueOf(receiver)
	}
	var proxyVal any
	if proxy == nil {
		proxyVal = reflect.Zero(reflect.TypeOf((*base.Proxy)(nil)).Elem())
	} else {
		proxyVal = reflect.ValueOf(proxy)
	}
	vals, err := c.Call("Exec", ctxVal, at, seq, receiverVal, proxyVal)
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}

func (c Cmd) Timeout() (timeout time.Duration) {
	vals, err := c.Call("Timeout")
	if err != nil {
		panic(err)
	}
	timeout = vals[0].(time.Duration)
	return
}
