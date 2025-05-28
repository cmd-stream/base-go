package mock

import (
	"context"
	"time"

	"github.com/cmd-stream/base-go"
	"github.com/ymz-ncnk/mok"
)

type ExecFn func(ctx context.Context, seq base.Seq, at time.Time, receiver any,
	proxy base.Proxy) (err error)
type TimeoutFn func() (timeout time.Duration)

func NewCmd() Cmd {
	return Cmd{mok.New("Cmd")}
}

type Cmd struct {
	*mok.Mock
}

func (c Cmd) RegisterExec(fn ExecFn) Cmd {
	c.Register("Exec", fn)
	return c
}

func (c Cmd) Exec(ctx context.Context, seq base.Seq, at time.Time, receiver any,
	proxy base.Proxy) (err error) {
	vals, err := c.Call("Exec", mok.SafeVal[context.Context](ctx), seq, at,
		mok.SafeVal[any](receiver), mok.SafeVal[base.Proxy](proxy))
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}
