package base

import (
	"context"
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
	Send(seq Seq, result Result) error
	SendWithDeadline(deadline time.Time, seq Seq, result Result) error
}

// Cmd defines the general Command interface.
//
// The Exec method is invoked by the server's Invoker. It executes the Command
// with the following parameters:
//   - ctx: Execution context.
//   - at: Timestamp when the server received the Command.
//   - seq: Sequence number assigned to the Command.
//   - receiver: An instance of type T that processes the Command.
//   - proxy: A server transport proxy used to send Results back to the client.
type Cmd[T any] interface {
	Exec(ctx context.Context, at time.Time, seq Seq, receiver T, proxy Proxy) error
}
