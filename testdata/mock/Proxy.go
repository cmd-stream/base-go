package mock

import (
	"net"
	"time"

	"github.com/cmd-stream/base-go"
	"github.com/ymz-ncnk/mok"
)

type SendFn func(seq base.Seq, result base.Result) (n int, err error)
type SendWithDeadlineFn func(seq base.Seq, result base.Result,
	deadline time.Time) (n int, err error)

func NewProxy() Proxy {
	return Proxy{mok.New("Proxy")}
}

type Proxy struct {
	*mok.Mock
}

func (p Proxy) RegisterLocalAddr(fn LocalAddrFn) Proxy {
	p.Register("LocalAddr", fn)
	return p
}

func (p Proxy) RegisterRemoteAddr(fn RemoteAddrFn) Proxy {
	p.Register("RemoteAddr", fn)
	return p
}

func (p Proxy) RegisterSend(fn SendFn) Proxy {
	p.Register("Send", fn)
	return p
}

func (p Proxy) RegisterSendWithDeadline(fn SendWithDeadlineFn) Proxy {
	p.Register("SendWithDeadline", fn)
	return p
}

func (p Proxy) LocalAddr() (addr net.Addr) {
	vals, err := p.Call("LocalAddr")
	if err != nil {
		panic(err)
	}
	addr, _ = vals[0].(net.Addr)
	return
}

func (p Proxy) RemoteAddr() (addr net.Addr) {
	vals, err := p.Call("RemoteAddr")
	if err != nil {
		panic(err)
	}
	addr, _ = vals[0].(net.Addr)
	return
}

func (p Proxy) Send(seq base.Seq, result base.Result) (n int, err error) {
	vals, err := p.Call("Send", seq, result)
	if err != nil {
		panic(err)
	}
	n = vals[0].(int)
	err, _ = vals[1].(error)
	return
}

func (p Proxy) SendWithDeadline(seq base.Seq, result base.Result,
	deadline time.Time) (n int, err error) {
	vals, err := p.Call("SendWithDeadline", seq, result, deadline)
	if err != nil {
		panic(err)
	}
	n = vals[0].(int)
	err, _ = vals[1].(error)
	return
}
