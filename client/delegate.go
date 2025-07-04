package client

import (
	"net"
	"sync"
	"time"

	base "github.com/cmd-stream/core-go"
)

// Delegate helps the client to send Commands and receive Results.
type Delegate[T any] interface {
	LocalAddr() net.Addr
	RemoteAddr() net.Addr

	SetSendDeadline(deadline time.Time) error
	Send(seq base.Seq, cmd base.Cmd[T]) (n int, err error)
	Flush() error

	SetReceiveDeadline(deadline time.Time) error
	Receive() (seq base.Seq, result base.Result, n int, err error)

	Close() error
}

// KeepaliveDelegate defines the Keepalive method.
//
// This delegate can be used if you want the client to keepalive connection to
// the server.
type KeepaliveDelegate[T any] interface {
	Delegate[T]
	Keepalive(muSn *sync.Mutex)
}

// ReconnectDelegate defines the Reconnect method.
//
// This delegate can be used if you want the client to reconnect to the server
// in case of a connection loss.
type ReconnectDelegate[T any] interface {
	Delegate[T]
	Reconnect() error
}
