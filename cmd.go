package base

import (
	"context"
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
	Send(seq Seq, result Result) error
	SendWithDeadline(deadline time.Time, seq Seq, result Result) error
}

// Cmd represents the general Command interface.
//
// Exec method is used by the Invoker on the server. Parameters:
//   - ctx: Execution context.
//   - at: Timestamp when the server received the Command.
//   - seq: Sequence number assigned to the Command.
//   - receiver: The Receiver of type T.
//   - proxy: Proxy of the server transport, used for sending Results back
//     to the client.
type Cmd[T any] interface {
	Exec(ctx context.Context, at time.Time, seq Seq, receiver T, proxy Proxy) error
}
