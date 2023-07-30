package base

import (
	"net"
	"time"
)

// Listener represents a network listener for the cmd-stream server.
//
// On Close it should not close already accepted connections.
type Listener interface {
	Addr() net.Addr
	SetDeadline(time.Time) error
	Accept() (net.Conn, error)
	Close() error
}
