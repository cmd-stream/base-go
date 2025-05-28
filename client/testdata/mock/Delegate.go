package bcmock

import (
	"net"
	"time"

	"github.com/cmd-stream/base-go"
	"github.com/ymz-ncnk/mok"
)

type LocalAddrFn func() (addr net.Addr)
type RemoteAddrFn func() (addr net.Addr)
type SetSendDeadlineFn func(deadline time.Time) (err error)
type SendFn func(seq base.Seq, cmd base.Cmd[any]) (n int, err error)
type FlushFn func() (err error)
type SetReceiveDeadlineFn func(deadline time.Time) (err error)
type ReceiveFn func() (seq base.Seq, result base.Result, n int, err error)
type CloseFn func() (err error)

func NewDelegate() Delegate {
	return Delegate{Mock: mok.New("Delegate")}
}

type Delegate struct {
	*mok.Mock
}

func (m Delegate) RegisterLocalAddr(fn LocalAddrFn) Delegate {
	m.Register("LocalAddr", fn)
	return m
}

func (m Delegate) RegisterRemoteAddr(fn RemoteAddrFn) Delegate {
	m.Register("RemoteAddr", fn)
	return m
}

func (m Delegate) RegisterNSetSendDeadline(n int, fn SetSendDeadlineFn) Delegate {
	m.RegisterN("SetSendDeadline", n, fn)
	return m
}

func (m Delegate) RegisterSetSendDeadline(fn SetSendDeadlineFn) Delegate {
	m.Register("SetSendDeadline", fn)
	return m
}

func (m Delegate) RegisterNSend(n int, fn SendFn) Delegate {
	m.RegisterN("Send", n, fn)
	return m
}

func (m Delegate) RegisterSend(fn SendFn) Delegate {
	m.Register("Send", fn)
	return m
}

func (m Delegate) RegisterNFlush(n int, fn FlushFn) Delegate {
	m.RegisterN("Flush", n, fn)
	return m
}

func (m Delegate) RegisterFlush(fn FlushFn) Delegate {
	m.Register("Flush", fn)
	return m
}

func (m Delegate) RegisterNSetReceiveDeadline(n int,
	fn SetReceiveDeadlineFn) Delegate {
	m.RegisterN("SetReceiveDeadline", n, fn)
	return m
}

func (m Delegate) RegisterSetReceiveDeadline(fn SetReceiveDeadlineFn) Delegate {
	m.Register("SetReceiveDeadline", fn)
	return m
}

func (m Delegate) RegisterReceive(fn ReceiveFn) Delegate {
	m.Register("Receive", fn)
	return m
}

func (m Delegate) RegisterClose(fn CloseFn) Delegate {
	m.Register("Close", fn)
	return m
}

func (m Delegate) LocalAddr() (addr net.Addr) {
	vals, err := m.Call("LocalAddr")
	if err != nil {
		panic(err)
	}
	addr, _ = vals[0].(net.Addr)
	return
}

func (m Delegate) RemoteAddr() (addr net.Addr) {
	vals, err := m.Call("RemoteAddr")
	if err != nil {
		panic(err)
	}
	addr, _ = vals[0].(net.Addr)
	return
}

func (m Delegate) SetSendDeadline(deadline time.Time) (err error) {
	vals, err := m.Call("SetSendDeadline", deadline)
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}

func (m Delegate) Send(seq base.Seq, cmd base.Cmd[any]) (n int, err error) {
	vals, err := m.Call("Send", seq, mok.SafeVal[base.Cmd[any]](cmd))
	if err != nil {
		panic(err)
	}
	n = vals[0].(int)
	err, _ = vals[1].(error)
	return
}

func (m Delegate) Flush() (err error) {
	vals, err := m.Call("Flush")
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}

func (m Delegate) SetReceiveDeadline(deadline time.Time) (err error) {
	vals, err := m.Call("SetReceiveDeadline", deadline)
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}

func (m Delegate) Receive() (seq base.Seq, result base.Result, n int,
	err error) {
	vals, err := m.Call("Receive")
	if err != nil {
		panic(err)
	}
	seq = vals[0].(base.Seq)
	result, _ = vals[1].(base.Result)
	n = vals[2].(int)
	err, _ = vals[3].(error)
	return
}

func (m Delegate) Close() (err error) {
	vals, err := m.Call("Close")
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}
