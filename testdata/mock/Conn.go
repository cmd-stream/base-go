package mock

import (
	"net"
	"time"

	"github.com/ymz-ncnk/mok"
)

type LocalAddrFn func() (addr net.Addr)
type RemoteAddrFn func() (addr net.Addr)
type SetDeadlineFn func(deadline time.Time) (err error)
type SetReadDeadlineFn func(deadline time.Time) (err error)
type SetWriteDeadlineFn func(deadline time.Time) (err error)
type WriteFn func(b []byte) (n int, err error)
type ReadFn func(b []byte) (n int, err error)
type CloseFn func() (err error)

// NewConn creates a new Conn.
func NewConn() Conn {
	return Conn{mok.New("Conn")}
}

type Conn struct {
	*mok.Mock
}

func (m Conn) RegisterLocalAddr(fn LocalAddrFn) Conn {
	m.Register("LocalAddr", fn)
	return m
}

func (m Conn) RegisterRead(fn ReadFn) Conn {
	m.Register("Read", fn)
	return m
}

func (m Conn) RegisterNRead(n int, fn ReadFn) Conn {
	m.RegisterN("Read", n, fn)
	return m
}

func (m Conn) RegisterRemoteAddr(fn RemoteAddrFn) Conn {
	m.Register("RemoteAddr", fn)
	return m
}

func (m Conn) RegisterSetDeadline(fn SetDeadlineFn) Conn {
	m.Register("SetDeadline", fn)
	return m
}

func (m Conn) RegisterNSetReadDeadline(n int, fn SetReadDeadlineFn) Conn {
	m.RegisterN("SetReadDeadline", n, fn)
	return m
}

func (m Conn) RegisterSetReadDeadline(fn SetReadDeadlineFn) Conn {
	m.Register("SetReadDeadline", fn)
	return m
}

func (m Conn) RegisterSetWriteDeadline(fn SetWriteDeadlineFn) Conn {
	m.Register("SetWriteDeadline", fn)
	return m
}

func (m Conn) RegisterWrite(fn WriteFn) Conn {
	m.Register("Write", fn)
	return m
}

func (m Conn) RegisterClose(fn CloseFn) Conn {
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
