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
// execution could be time-limited. This can be achieved by using the at
// parameter, which indicates when the command was received
//
//		deadline := at.Add(CmdTimeout)
//		ownCtx, cancel := context.WithDeadline(ctx, deadline)
//		// Do the context dependent work.
//		...
//		err = proxy.SendWithDeadline(deadline, seq, result)
//	  ...
//
// With Proxy you can send more than one result back, the last one should have
// result.LastOne() == true.
type Cmd[T any] interface {
	Exec(ctx context.Context, at time.Time, seq Seq, receiver T, proxy Proxy) error
}
