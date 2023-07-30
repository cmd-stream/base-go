package mock

import (
	"net"
	"reflect"
	"time"

	"github.com/cmd-stream/base-go"
	"github.com/ymz-ncnk/mok"
)

func NewReconnectClientDelegate() ReconnectClientDelegate {
	return ReconnectClientDelegate{
		Mock: mok.New("ReconnectClientDelegate"),
	}
}

type ReconnectClientDelegate struct {
	*mok.Mock
}

func (m ReconnectClientDelegate) RegisterReconnect(
	fn func() (err error)) ReconnectClientDelegate {
	m.Register("Reconnect", fn)
	return m
}

func (m ReconnectClientDelegate) RegisterLocalAddr(
	fn func() (addr net.Addr)) ReconnectClientDelegate {
	m.Register("LocalAddr", fn)
	return m
}

func (m ReconnectClientDelegate) RegisterRemoteAddr(
	fn func() (addr net.Addr)) ReconnectClientDelegate {
	m.Register("RemoteAddr", fn)
	return m
}

func (m ReconnectClientDelegate) RegisterNSetSendDeadline(n int,
	fn func(deadline time.Time) (err error)) ReconnectClientDelegate {
	m.RegisterN("SetSendDeadline", n, fn)
	return m
}

func (m ReconnectClientDelegate) RegisterSetSendDeadline(
	fn func(deadline time.Time) (err error)) ReconnectClientDelegate {
	m.Register("SetSendDeadline", fn)
	return m
}

func (m ReconnectClientDelegate) RegisterNSend(n int,
	fn func(seq base.Seq, cmd base.Cmd[any]) (err error)) ReconnectClientDelegate {
	m.RegisterN("Send", n, fn)
	return m
}

func (m ReconnectClientDelegate) RegisterSend(
	fn func(seq base.Seq, cmd base.Cmd[any]) (err error)) ReconnectClientDelegate {
	m.Register("Send", fn)
	return m
}

func (m ReconnectClientDelegate) RegisterNFlush(n int,
	fn func() (err error)) ReconnectClientDelegate {
	m.RegisterN("Flush", n, fn)
	return m
}

func (m ReconnectClientDelegate) RegisterFlush(
	fn func() (err error)) ReconnectClientDelegate {
	m.Register("Flush", fn)
	return m
}

func (m ReconnectClientDelegate) RegisterNSetReceiveDeadline(n int,
	fn func(deadline time.Time) (err error)) ReconnectClientDelegate {
	m.RegisterN("SetReceiveDeadline", n, fn)
	return m
}

func (m ReconnectClientDelegate) RegisterSetReceiveDeadline(
	fn func(deadline time.Time) (err error)) ReconnectClientDelegate {
	m.Register("SetReceiveDeadline", fn)
	return m
}

func (m ReconnectClientDelegate) RegisterReceive(
	fn func() (seq base.Seq, result base.Result, err error)) ReconnectClientDelegate {
	m.Register("Receive", fn)
	return m
}

func (m ReconnectClientDelegate) RegisterClose(
	fn func() (err error)) ReconnectClientDelegate {
	m.Register("Close", fn)
	return m
}

func (m ReconnectClientDelegate) Reconnect() (err error) {
	vals, err := m.Call("Reconnect")
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}

func (m ReconnectClientDelegate) LocalAddr() (addr net.Addr) {
	vals, err := m.Call("LocalAddr")
	if err != nil {
		panic(err)
	}
	addr, _ = vals[0].(net.Addr)
	return
}

func (m ReconnectClientDelegate) RemoteAddr() (addr net.Addr) {
	vals, err := m.Call("RemoteAddr")
	if err != nil {
		panic(err)
	}
	addr, _ = vals[0].(net.Addr)
	return
}

func (m ReconnectClientDelegate) SetSendDeadline(deadline time.Time) (err error) {
	vals, err := m.Call("SetSendDeadline", deadline)
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}

func (m ReconnectClientDelegate) Send(seq base.Seq, cmd base.Cmd[any]) (err error) {
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

func (m ReconnectClientDelegate) Flush() (err error) {
	vals, err := m.Call("Flush")
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}

func (m ReconnectClientDelegate) SetReceiveDeadline(deadline time.Time) (err error) {
	vals, err := m.Call("SetReceiveDeadline", deadline)
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}

func (m ReconnectClientDelegate) Receive() (seq base.Seq, result base.Result, err error) {
	vals, err := m.Call("Receive")
	if err != nil {
		panic(err)
	}
	seq, _ = vals[0].(base.Seq)
	result, _ = vals[1].(base.Result)
	err, _ = vals[2].(error)
	return
}

func (m ReconnectClientDelegate) Close() (err error) {
	vals, err := m.Call("Close")
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}

// func NewReconnectClientDelegate() ReconnectClientDelegate {
// 	return ReconnectClientDelegate{
// 		ClientDelegate{
// 			Mock: mok.New("ReconnectClientDelegate"),
// 		},
// 	}
// }

// type ReconnectClientDelegate struct {
// 	ClientDelegate
// }

// func (m ReconnectClientDelegate) RegisterReconnect(
// 	fn func(seq base.Seq, cmd base.Cmd[any]) (err error)) ReconnectClientDelegate {
// 	m.Register("Reconnect", fn)
// 	return m
// }

// func (m ReconnectClientDelegate) Reconnect() (err error) {
// 	vals, err := m.Call("Reconnect")
// 	if err != nil {
// 		panic(err)
// 	}
// 	err, _ = vals[0].(error)
// 	return
// }
