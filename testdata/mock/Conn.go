package mock

import (
	"net"
	"time"

	"github.com/ymz-ncnk/mok"
)

// NewConn creates a new Conn.
func NewConn() Conn {
	return Conn{
		Mock: mok.New("Conn"),
	}
}

type Conn struct {
	*mok.Mock
}

func (m Conn) RegisterLocalAddr(
	fn func() (addr net.Addr)) Conn {
	m.Register("LocalAddr", fn)
	return m
}

func (m Conn) RegisterNRead(n int,
	fn func(b []byte) (n int, err error)) Conn {
	m.RegisterN("Read", n, fn)
	return m
}

func (m Conn) RegisterRead(
	fn func(b []byte) (n int, err error)) Conn {
	m.Register("Read", fn)
	return m
}

func (m Conn) RegisterRemoteAddr(
	fn func() (addr net.Addr)) Conn {
	m.Register("RemoteAddr", fn)
	return m
}

func (m Conn) RegisterSetDeadline(
	fn func(deadline time.Time) (err error)) Conn {
	m.Register("SetDeadline", fn)
	return m
}

func (m Conn) RegisterNSetReadDeadline(n int,
	fn func(deadline time.Time) (err error)) Conn {
	m.RegisterN("SetReadDeadline", n, fn)
	return m
}

func (m Conn) RegisterSetReadDeadline(
	fn func(deadline time.Time) (err error)) Conn {
	m.Register("SetReadDeadline", fn)
	return m
}

func (m Conn) RegisterSetWriteDeadline(
	fn func(deadline time.Time) (err error)) Conn {
	m.Register("SetWriteDeadline", fn)
	return m
}

func (m Conn) RegisterWrite(
	fn func(b []byte) (n int, err error)) Conn {
	m.Register("Write", fn)
	return m
}

func (m Conn) RegisterClose(
	fn func() (err error)) Conn {
	m.Register("Close", fn)
	return m
}

func (m Conn) LocalAddr() (addr net.Addr) {
	vals, err := m.Call("LocalAddr")
	if err != nil {
		panic(err)
	}
	addr, _ = vals[0].(net.Addr)
	return
}

func (m Conn) RemoteAddr() (addr net.Addr) {
	vals, err := m.Call("RemoteAddr")
	if err != nil {
		panic(err)
	}
	addr, _ = vals[0].(net.Addr)
	return
}

func (m Conn) SetDeadline(deadline time.Time) (err error) {
	vals, err := m.Call("SetDeadline", deadline)
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}

func (m Conn) SetReadDeadline(deadline time.Time) (err error) {
	vals, err := m.Call("SetReadDeadline", deadline)
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}

func (m Conn) SetWriteDeadline(deadline time.Time) (err error) {
	vals, err := m.Call("SetWriteDeadline", deadline)
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}

func (m Conn) Write(b []byte) (n int, err error) {
	vals, err := m.Call("Write", b)
	if err != nil {
		panic(err)
	}
	n = vals[0].(int)
	err, _ = vals[1].(error)
	return
}

func (m Conn) Read(b []byte) (n int, err error) {
	vals, err := m.Call("Read", b)
	if err != nil {
		panic(err)
	}
	n = vals[0].(int)
	err, _ = vals[1].(error)
	return
}

func (m Conn) Close() (err error) {
	vals, err := m.Call("Close")
	if err != nil {
		panic(err)
	}
	err, _ = vals[0].(error)
	return
}
