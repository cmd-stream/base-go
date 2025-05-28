package base

import (
	"net"
	"time"
)

// Seq represents the sequence number of a Command.
//
// The sequence number ensures that each Command can be uniquely identified and
// mapped to its corresponding Results.
type Seq int64

// Proxy represents a server transport proxy, enabling Commands to send Results
// back.
//
// Implementation of this interface must be thread-safe.
type Proxy interface {
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Send(seq Seq, result Result) (n int, err error)
	SendWithDeadline(seq Seq, result Result, deadline time.Time) (n int, err error)
}
