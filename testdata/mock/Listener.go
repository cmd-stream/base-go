package mock

import (
	"net"
	"time"

	"github.com/ymz-ncnk/mok"
)

func NewListener() Listener {
	return Listener{
		Mock: mok.New("Listener"),
	}
}

type Listener struct {
	*mok.Mock
}

func (m Listener) RegisterAddr(
	fn func() (addr net.Addr)) Listener {
	m.Register("Addr", fn)
	return m
}

func (m Listener) RegisterSetDeadline(
	fn func(deadline time.Time) (err error)) Listener {
	m.Register("SetDeadline", fn)
	return m
}

func (m Listener) RegisterNSetDeadline(n int,
	fn func(deadline time.Time) (err error)) Listener {
	m.RegisterN("SetDeadline", n, fn)
	return m
}

func (m Listener) RegisterAccept(
	fn func() (conn net.Conn, err error)) Listener {
	m.Register("Accept", fn)
	return m
}

func (m Listener) RegisterClose(
	fn func() (err error)) Listener {
	m.Register("Close", fn)
	return m
}

func (m Listener) Addr() (addr net.Addr) {
	vals, err := m.Call("Addr")
	if err != nil {
		panic(err)
	}
	addr, _ = vals[0].(net.Addr)
	return
}

func (m Listener) SetDeadline(deadline time.Time) (err error) {
	vals, err := m.Call("SetDeadline", deadline)
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}

func (m Listener) Accept() (conn net.Conn, err error) {
	vals, err := m.Call("Accept")
	if err != nil {
		panic(err)
	}
	conn, _ = vals[0].(net.Conn)
	err, _ = vals[1].(error)
	return
}

func (m Listener) Close() (conn error) {
	vals, err := m.Call("Close")
	if err != nil {
		panic(err)
	}
	conn, _ = vals[0].(error)
	return
}
