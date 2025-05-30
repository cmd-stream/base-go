package csrv

import (
	"context"
	"net"
)

// NewWorker creates a new Worker.
func NewWorker(conns <-chan net.Conn, delegate Delegate,
	callback LostConnCallback) Worker {
	ctx, cancel := context.WithCancel(context.Background())
	return Worker{ctx, cancel, conns, delegate, callback}
}

// Worker is a server worker.
//
// It receives connections from the conns channel one at a time and processes
// them using the Delegate. Additionally, it implements jointwork.Task, allowing
// it to work in conjunction with ConnReceiver.
//
// If connection processing fails with an error, the error is passed to
// LostConnCallback.
type Worker struct {
	ctx      context.Context
	cancel   context.CancelFunc
	conns    <-chan net.Conn
	delegate Delegate
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
