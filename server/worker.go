package server

import (
	"context"
	"net"

	"github.com/cmd-stream/base-go"
)

// NewWorker creates a new Worker.
func NewWorker(conns <-chan net.Conn, delegate base.ServerDelegate,
	callback LostConnCallback,
) Worker {
	ctx, cancel := context.WithCancel(context.Background())
	return Worker{ctx, cancel, conns, delegate, callback}
}

// Worker is a Server worker.
//
// It receives a connections from the conns channel and handles them using the
// delegate. Also it implements jointwork.Task, so it + ConnReceiver may do the
// job together.
//
// If the connection processing failed with an error, it is passed to the
// LostConnCallback.
type Worker struct {
	ctx      context.Context
	cancel   context.CancelFunc
	conns    <-chan net.Conn
	delegate base.ServerDelegate
	callback LostConnCallback
}

func (w Worker) Run() (err error) {
	var (
		conn net.Conn
		more bool
	)
	for {
		select {
		case <-w.ctx.Done():
			return ErrClosed
		case conn, more = <-w.conns:
			if !more {
				return nil
			}
			err = w.delegate.Handle(w.ctx, conn)
			if err != nil {
				if err == context.Canceled {
					err = ErrClosed
				}
				if w.callback != nil {
					w.callback(conn.RemoteAddr(), err)
				}
				if err == ErrClosed {
					return
				}
			}
		}
	}
}

func (w Worker) Stop() (err error) {
	w.cancel()
	return
}
