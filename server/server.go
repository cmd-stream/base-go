package bsrv

import (
	"net"
	"sync"

	"github.com/cmd-stream/base-go"
	"github.com/ymz-ncnk/jointwork-go"
)

// WorkersCount defines the default number of workers.
const WorkersCount = 8

// LostConnCallback is invoked when the server loses its connection to a client.
type LostConnCallback = func(addr net.Addr, err error)

// New creates a new server.
func New(delegate Delegate, ops ...SetOption) (s *Server) {
	s = &Server{delegate: delegate, options: Options{WorkersCount: WorkersCount}}
	Apply(ops, &s.options)
	return
}

// Server represents a cmd-stream server.
//
// It utilizes a configurable number of Workers to manage client connections
// using a specified ServerDelegate.
type Server struct {
	delegate Delegate
	receiver *ConnReceiver
	mu       sync.Mutex
	options  Options
}

// Serve accepts and processes incoming connections on the listener using
// Workers.
//
// Each worker handles one connection at a time.
//
// This function always returns a non-nil error:
//   - If Conf.WorkersCount == 0, it returns ErrNoWorkers.
//   - If the server was shut down, it returns ErrShutdown.
//   - If the server was closed, it returns ErrClosed.
func (s *Server) Serve(listener base.Listener) (err error) {
	if s.options.WorkersCount <= 0 {
		return ErrNoWorkers
	}
	conns := make(chan net.Conn, s.options.WorkersCount)
	s.setReceiver(listener, conns)
	var (
		tasks = s.makeTasks(conns, s.delegate)
		jw    = jointwork.New(tasks)
	)
	if err = jw.Run(); err == nil {
		return ErrShutdown
	}
	firstErr := err.(interface{ Get(i int) error }).Get(0)
	return firstErr.(*jointwork.TaskError).Cause()
}

// Shutdown stops the server from receiving new connections.
//
// If server is not serving returns ErrNotServing.
func (s *Server) Shutdown() (err error) {
	if !s.serving() {
		return ErrNotServing
	}
	return s.receiver.Shutdown()
}

// Close closes the server, all existing connections will be closed.
//
// If server is not serving returns ErrNotServing.
func (s *Server) Close() (err error) {
	if !s.serving() {
		return ErrNotServing
	}
	return s.receiver.Stop()
}

func (s *Server) setReceiver(listener base.Listener, conns chan net.Conn) {
	s.mu.Lock()
	s.receiver = NewConnReceiver(listener, conns, s.options.ConnReceiver...)
	s.mu.Unlock()
}

func (s *Server) makeTasks(conns chan net.Conn, delegate Delegate) (
	tasks []jointwork.Task) {
	tasks = make([]jointwork.Task, 1+s.options.WorkersCount)
	tasks[0] = s.receiver
	for i := 1; i < len(tasks); i++ {
		tasks[i] = NewWorker(conns, delegate, s.options.LostConnCallback)
	}
	return
}

func (s *Server) serving() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.receiver != nil
}
