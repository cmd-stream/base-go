package server

import (
	"net"
	"sync"
	"time"

	"github.com/cmd-stream/base-go"
)

const (
	inProgress int = iota
	shutdown
	closed
)

// NewConnReceiver creates a new ConnReceiver.
func NewConnReceiver(conf ConnReceiverConf, listener base.Listener,
	conns chan net.Conn) *ConnReceiver {
	return &ConnReceiver{
		conf:     conf,
		listener: listener,
		conns:    conns,
		stopped:  make(chan struct{}),
	}
}

// ConnReceiver accepts incoming connections on the listener and adds them to
// the conns channel.
//
// It can wait for the first connection for a limited amount of time, after
// which, it stops. Also ConnReceiver implements the jointwork.Task interface,
// so it + Workers may do the job together.
type ConnReceiver struct {
	conf     ConnReceiverConf
	listener base.Listener
	conns    chan net.Conn
	state    int
	stopped  chan struct{}
	mu       sync.Mutex
}

func (r *ConnReceiver) Run() (err error) {
	defer func() {
		r.postRun()
	}()
	if err = r.acceptFirstConn(); err != nil {
		return r.correctErr(err)
	}
	return r.correctErr(r.acceptConns())
}

// Shutdown stops the receiver - the Run() method returns nil, it allows Workers
// to finish their work.
func (r *ConnReceiver) Shutdown() (err error) {
	return r.terminate(shutdown)
}

// Stop stops the receiver - the Run() method returns ErrClosed.
func (r *ConnReceiver) Stop() (err error) {
	return r.terminate(closed)
}

func (r *ConnReceiver) acceptFirstConn() (err error) {
	if r.conf.FirstConnTimeout != 0 {
		defer func() {
			if err == nil {
				err = r.listener.SetDeadline(time.Time{})
			}
		}()
		err = r.listener.SetDeadline(time.Now().Add(r.conf.FirstConnTimeout))
		if err != nil {
			return err
		}
	}
	conn, err := r.listener.Accept()
	if err != nil {
		return err
	}
	return r.queueConn(conn)
}

func (r *ConnReceiver) acceptConns() (err error) {
	var conn net.Conn
	for {
		conn, err = r.listener.Accept()
		if err != nil {
			return
		}
		if err = r.queueConn(conn); err != nil {
			return
		}
	}
}

func (r *ConnReceiver) queueConn(conn net.Conn) error {
	select {
	case <-r.stopped:
		if err := conn.Close(); err != nil {
			panic(err)
		}
		return ErrClosed
	case r.conns <- conn:
		return nil
	}
}

func (r *ConnReceiver) terminate(state int) (err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.state == inProgress {
		r.state = state
		if err = r.listener.Close(); err != nil {
			r.state = inProgress
			return
		}
		close(r.stopped)
	}
	return
}

func (r *ConnReceiver) correctErr(err error) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	switch r.state {
	case inProgress:
		return err
	case shutdown:
		return nil
	case closed:
		return ErrClosed
	default:
		panic("unexpected state")
	}
}

func (r *ConnReceiver) postRun() {
	r.mu.Lock()
	defer r.mu.Unlock()
	close(r.conns)
	if r.state == shutdown {
		return
	}
	for conn := range r.conns {
		if err := conn.Close(); err != nil {
			panic(err)
		}
	}
}
