package base

import (
	"context"
	"net"
	"sync"
	"time"
)

// ClientReconnectDelegate defines the Reconnect method.
//
// This delegate can be used if you want the Client to reconnect to the Server
// in case of a connection loss.
type ClientReconnectDelegate[T any] interface {
	ClientDelegate[T]
	Reconnect() error
}

// ClientDelegate helps the client to send commands and receive results.
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

type ClientKeepaliveDelegate[T any] interface {
	ClientDelegate[T]
	Keepalive(muSn *sync.Mutex)
}

// ServerDelegate helps the server handle incoming connections.
//
// Handle method should return the context.Canceled error if a context is done.
type ServerDelegate interface {
	Handle(ctx context.Context, conn net.Conn) error
}
