package mock

import (
	"net"
	"reflect"
	"time"

	"github.com/cmd-stream/base-go"
	"github.com/ymz-ncnk/mok"
)

func NewClientDelegate() ClientDelegate {
	return ClientDelegate{
		Mock: mok.New("ClientDelegate"),
	}
}

type ClientDelegate struct {
	*mok.Mock
}

func (m ClientDelegate) RegisterLocalAddr(
	fn func() (addr net.Addr)) ClientDelegate {
	m.Register("LocalAddr", fn)
	return m
}

func (m ClientDelegate) RegisterRemoteAddr(
	fn func() (addr net.Addr)) ClientDelegate {
	m.Register("RemoteAddr", fn)
	return m
}

func (m ClientDelegate) RegisterNSetSendDeadline(n int,
	fn func(deadline time.Time) (err error)) ClientDelegate {
	m.RegisterN("SetSendDeadline", n, fn)
	return m
}

func (m ClientDelegate) RegisterSetSendDeadline(
	fn func(deadline time.Time) (err error)) ClientDelegate {
	m.Register("SetSendDeadline", fn)
	return m
}

func (m ClientDelegate) RegisterNSend(n int,
	fn func(seq base.Seq, cmd base.Cmd[any]) (err error)) ClientDelegate {
	m.RegisterN("Send", n, fn)
	return m
}

func (m ClientDelegate) RegisterSend(
	fn func(seq base.Seq, cmd base.Cmd[any]) (err error)) ClientDelegate {
	m.Register("Send", fn)
	return m
}

func (m ClientDelegate) RegisterNFlush(n int,
	fn func() (err error)) ClientDelegate {
	m.RegisterN("Flush", n, fn)
	return m
}

func (m ClientDelegate) RegisterFlush(
	fn func() (err error)) ClientDelegate {
	m.Register("Flush", fn)
	return m
}

func (m ClientDelegate) RegisterNSetReceiveDeadline(n int,
	fn func(deadline time.Time) (err error)) ClientDelegate {
	m.RegisterN("SetReceiveDeadline", n, fn)
	return m
}

func (m ClientDelegate) RegisterSetReceiveDeadline(
	fn func(deadline time.Time) (err error)) ClientDelegate {
	m.Register("SetReceiveDeadline", fn)
	return m
}

func (m ClientDelegate) RegisterReceive(
	fn func() (seq base.Seq, result base.Result, err error)) ClientDelegate {
	m.Register("Receive", fn)
	return m
}

func (m ClientDelegate) RegisterClose(
	fn func() (err error)) ClientDelegate {
	m.Register("Close", fn)
	return m
}

func (m ClientDelegate) LocalAddr() (addr net.Addr) {
	vals, err := m.Call("LocalAddr")
	if err != nil {
		panic(err)
	}
	addr, _ = vals[0].(net.Addr)
	return
}

func (m ClientDelegate) RemoteAddr() (addr net.Addr) {
	vals, err := m.Call("RemoteAddr")
	if err != nil {
		panic(err)
	}
	addr, _ = vals[0].(net.Addr)
	return
}

func (m ClientDelegate) SetSendDeadline(deadline time.Time) (err error) {
	vals, err := m.Call("SetSendDeadline", deadline)
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}

func (m ClientDelegate) Send(seq base.Seq, cmd base.Cmd[any]) (err error) {
	var cmdVal reflect.Value
	if cmd == nil {
		cmdVal = reflect.Zero(reflect.TypeOf((*base.Cmd[any])(nil)).Elem())
	} else {
		cmdVal = reflect.ValueOf(cmd)
	}
	vals, err := m.Call("Send", seq, cmdVal)
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}

func (m ClientDelegate) Flush() (err error) {
	vals, err := m.Call("Flush")
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}

func (m ClientDelegate) SetReceiveDeadline(deadline time.Time) (err error) {
	vals, err := m.Call("SetReceiveDeadline", deadline)
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}

func (m ClientDelegate) Receive() (seq base.Seq, result base.Result, err error) {
	vals, err := m.Call("Receive")
	if err != nil {
		panic(err)
	}
	seq, _ = vals[0].(base.Seq)
	result, _ = vals[1].(base.Result)
	err, _ = vals[2].(error)
	return
}

func (m ClientDelegate) Close() (err error) {
	vals, err := m.Call("Close")
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}
