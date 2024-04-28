package mock

import (
	"context"
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
	vals, err := c.Call("Exec", mok.SafeVal[context.Context](ctx), at, seq,
		mok.SafeVal[any](receiver), mok.SafeVal[base.Proxy](proxy))
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
