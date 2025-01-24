package base

import (
	"context"
	"net"
	"sync"
	"time"
)

// ClientDelegate helps the client to send Commands and receive Results.
type ClientDelegate[T any] interface {
	LocalAddr() net.Addr
	RemoteAddr() net.Addr

	SetSendDeadline(deadline time.Time) error
	Send(seq Seq, cmd Cmd[T]) error
	Flush() error

	SetReceiveDeadline(deadline time.Time) error
	Receive() (seq Seq, result Result, err error)

	Close() error
}

// ClientReconnectDelegate defines the Reconnect method.
//
// This delegate can be used if you want the client to reconnect to the server
// in case of a connection loss.
type ClientReconnectDelegate[T any] interface {
	ClientDelegate[T]
	Reconnect() error
}

// ClientKeepaliveDelegate defines the Keepalive method.
//
// This delegate can be used if you want the client to keepalive connection to
// the server.
type ClientKeepaliveDelegate[T any] interface {
	ClientDelegate[T]
	Keepalive(muSn *sync.Mutex)
}

// ServerDelegate helps the server handle incoming connections.
//
// Handle method should return context.Canceled error if the context is done.
type ServerDelegate interface {
	Handle(ctx context.Context, conn net.Conn) error
}
