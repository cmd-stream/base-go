package server

import (
	"net"
	"sync"

	"github.com/cmd-stream/base-go"
	"github.com/ymz-ncnk/jointwork-go"
)

// LostConnCallback is invoked when the Server loses its connection to a Client.
type LostConnCallback = func(addr net.Addr, err error)

// Server represents a cmd-stream server.
//
// It utilizes a configurable number of workers to manage client connections
// using a specified ServerDelegate.
type Server struct {
	Conf     Conf
	Delegate base.ServerDelegate
	Callback LostConnCallback
	receiver *ConnReceiver
	mu       sync.Mutex
}

// Serve accepts and process incoming connections on the listener using workers.
//
// Each worker can handle one connection at a time.
//
// Always returns a non-nil error. If Conf.WorkersCount == 0, returns
// ErrNoWorkers. If Server was shutdown returns ErrShutdown, if was closed -
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

// Shutdown stops the Server from receiving new connections.
//
// If Server is not serving returns ErrNotServing.
func (s *Server) Shutdown() (err error) {
	if !s.serving() {
		return ErrNotServing
	}
	return s.receiver.Shutdown()
}

// Close closes the Server, all existing connections will be closed.
//
// If Server is not serving returns ErrNotServing.
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
		tasks[i] = NewWorker(conns, delegate, s.Callback)
	}
	return
}

func (s *Server) serving() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.receiver != nil
}
