package server

import (
	"net"
	"sync"

	"github.com/cmd-stream/base-go"
	"github.com/ymz-ncnk/jointwork-go"
)

// Server is a cmd-stream server.
type Server struct {
	Conf     Conf
	Delegate base.ServerDelegate
	receiver *ConnReceiver
	mu       sync.Mutex
}

// Serve accepts incoming connections on the listener and processes them using
// the configured number of workers.
//
// Each worker can handle one connection at a time using a delegate.
//
// Always returns a non-nil error. If Conf.WorkersCount == 0, returns
// ErrNoWorkers. If the server was shutdown returns ErrShutdown, if was closed -
// ErrClosed.
func (s *Server) Serve(listener base.Listener) (err error) {
	if s.Conf.WorkersCount <= 0 {
		return ErrNoWorkers
	}
	conns := make(chan net.Conn, s.Conf.WorkersCount)
	s.setUpReceiver(listener, conns)
	var (
		tasks = s.makeTasks(conns, s.Delegate)
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
// If the server is not serving returns ErrNotServing.
func (s *Server) Shutdown() (err error) {
	if !s.serving() {
		return ErrNotServing
	}
	return s.receiver.Shutdown()
}

// Close closes the server, all existing connections will be closed.
//
// If the server is not serving returns ErrNotServing.
func (s *Server) Close() (err error) {
	if !s.serving() {
		return ErrNotServing
	}
	return s.receiver.Stop()
}

func (s *Server) setUpReceiver(listener base.Listener, conns chan net.Conn) {
	s.mu.Lock()
	s.receiver = NewConnReceiver(s.Conf.ConnReceiverConf, listener, conns)
	s.mu.Unlock()
}

func (s *Server) makeTasks(conns chan net.Conn, delegate base.ServerDelegate) (
	tasks []jointwork.Task) {
	tasks = make([]jointwork.Task, 1+s.Conf.WorkersCount)
	tasks[0] = s.receiver
	for i := 1; i < len(tasks); i++ {
		tasks[i] = NewWorker(conns, delegate, s.Conf.LostConnCallback)
	}
	return
}

func (s *Server) serving() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.receiver != nil
}
