package base

import (
	"context"
	"time"
)

// Seq represents a sequence number of the command.
//
// The sequence number is used to determine the mapping between the command and
// its results.
type Seq int64

// Proxy is a server transport proxy, it allows commands to send results back.
//
// Implementation of this interface should be thread-safe.
type Proxy interface {
	Send(seq Seq, result Result) error
	SendWithDeadline(deadline time.Time, seq Seq, result Result) error
}

// Cmd represents a cmd-stream command.
//
// All user-defined commands must implement this interface. The command
// execution could be time-limited. You can do this using the at parameter,
// which defines a time when the command was received:
//
//	deadline := at.Add(CmdTimeout)
//	ownCtx, cancel := context.WithDeadline(ctx, deadline)
//	// Do the context dependent work.
//	...
//	proxy.SendWithDeadline(deadline, seq, result)
//
// With proxy you can send more than one result back, the last one should have
// result.LastOne() == true.
type Cmd[T any] interface {
	Exec(ctx context.Context, at time.Time, seq Seq, receiver T, proxy Proxy) error
}
