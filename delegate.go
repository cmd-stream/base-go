package base

import (
	"context"
	"net"
	"time"
)

// ClientReconnectDelegate defines the Reconnect method.
//
// This delegate can be used if you want the client to reconnect to the server
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

// ServerDelegate helps the server handle incoming connections.
//
// If context is done, the Handle method should return context.Canceled error.
type ServerDelegate interface {
	Handle(ctx context.Context, conn net.Conn) error
}
