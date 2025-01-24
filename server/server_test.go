package server

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/cmd-stream/base-go"
	"github.com/cmd-stream/base-go/testdata/mock"
	"github.com/ymz-ncnk/mok"
)

const Delta = 100 * time.Millisecond

func TestServer(t *testing.T) {

	t.Run("Serving should fail if Conf.WorkersCount == 0", func(t *testing.T) {
		var (
			wantErr = ErrNoWorkers
			conf    = Conf{WorkersCount: 0}
			server  = &Server{Conf: conf}
			err     = server.Serve(nil)
		)
		if err != wantErr {
			t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
		}
	})

	t.Run("Server should be able to handle several connections",
		func(t *testing.T) {
			var (
				wantErr       = ErrClosed
				wantHandleErr = errors.New("handle conn failed")
				wg            = func() *sync.WaitGroup {
					wg := &sync.WaitGroup{}
					wg.Add(2)
					return wg
				}()
				addr1    = &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9900}
				conn1    = MakeConn(addr1)
				addr2    = &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9900}
				conn2    = MakeConn(addr2)
				delegate = mock.NewServerDelegate().RegisterNHandle(2,
					func(ctx context.Context, conn net.Conn) (err error) {
						// TODO
						return wantHandleErr
					},
				)
				conf = Conf{
					ConnReceiverConf: ConnReceiverConf{FirstConnTimeout: time.Second},
					WorkersCount:     2,
				}
				callback LostConnCallback = func(addr net.Addr, err error) {
					if err != wantHandleErr {
						t.Errorf("unexpected error, want '%v' actual '%v'", ErrClosed, err)
					}
					wg.Done()
				}
				listenerErr = errors.New("listener closed")
				listener    = func() mock.Listener {
					done := make(chan struct{})
					return mock.NewListener().RegisterNSetDeadline(2,
						func(t time.Time) error { return nil },
					).RegisterAccept(
						func() (net.Conn, error) { return conn1, nil },
					).RegisterAccept(
						func() (net.Conn, error) { return conn2, nil },
					).RegisterAccept(
						func() (net.Conn, error) {
							<-done
							return nil, listenerErr
						},
					).RegisterClose(
						func() error { close(done); return nil },
					)
					// .RegisterClose(func() (err error) { return nil })
				}()
				mocks = []*mok.Mock{conn1.Mock, conn2.Mock, listener.Mock}
			)
			server, errs := MakeServerAndServe(conf, listener, delegate, callback)
			wg.Wait()
			if err := server.Close(); err != nil {
				t.Fatal(err)
			}
			testAsyncErr(wantErr, errs, mocks, t)
		})

	t.Run("We should be able to shutdown the server after it receives a conn",
		func(t *testing.T) {
			var (
				wantErr         = ErrShutdown
				wantLostConnErr = errors.New("conn closed by client")
				wg              = func() *sync.WaitGroup {
					wg := &sync.WaitGroup{}
					wg.Add(2)
					return wg
				}()
				addr = &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9000}
				conn = MakeConn(addr)
				conf = Conf{
					WorkersCount: 1,
				}
				callback LostConnCallback = func(a net.Addr, err error) {
					if a != addr {
						t.Errorf("unexpected addr, want '%v' actual '%v'", addr, a)
					}
					if err != wantLostConnErr {
						t.Errorf("unexpected err, want '%v' actual '%v'",
							wantLostConnErr,
							err)
					}
				}
				listener = func() mock.Listener {
					listenerDone := make(chan struct{})
					return mock.NewListener().RegisterAccept(
						func() (net.Conn, error) { return conn, nil },
					).RegisterAccept(
						func() (net.Conn, error) {
							wg.Done()
							<-listenerDone
							return nil, errors.New("listener closed")
						},
					).RegisterClose(
						func() error { close(listenerDone); return nil },
					)
				}()
				delegate = mock.NewServerDelegate().RegisterHandle(
					func(ctx context.Context, conn net.Conn) (err error) {
						wg.Done()
						time.Sleep(100 * time.Millisecond)
						return wantLostConnErr
					},
				)
				mocks = []*mok.Mock{listener.Mock}
			)
			server, errs := MakeServerAndServe(conf, listener, delegate, callback)
			wg.Wait()
			if err := server.Shutdown(); err != nil {
				t.Fatal(err)
			}
			testAsyncErr(wantErr, errs, mocks, t)
		})

	t.Run("We should be able to close the server after it receives a conn",
		func(t *testing.T) {
			var (
				wantErr         = ErrClosed
				wantLostConnErr = ErrClosed
				wg              = func() *sync.WaitGroup {
					wg := &sync.WaitGroup{}
					wg.Add(2)
					return wg
				}()
				addr = &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9000}
				conn = MakeConn(addr)
				conf = Conf{
					WorkersCount: 1,
				}
				callback LostConnCallback = func(a net.Addr, err error) {
					if a != addr {
						t.Errorf("unexpected addr, want '%v' actual '%v'", addr, a)
					}
					if err != wantLostConnErr {
						t.Errorf("unexpected err, want '%v' actual '%v'",
							wantLostConnErr,
							err)
					}
				}
				listener = func() mock.Listener {
					listenerDone := make(chan struct{})
					return mock.NewListener().RegisterAccept(
						func() (net.Conn, error) { return conn, nil },
					).RegisterAccept(
						func() (net.Conn, error) {
							wg.Done()
							<-listenerDone
							return nil, errors.New("listener closed")
						},
					).RegisterClose(
						func() error { close(listenerDone); return nil },
					)
				}()
				delegate = mock.NewServerDelegate().RegisterHandle(
					func(ctx context.Context, conn net.Conn) (err error) {
						wg.Done()
						<-ctx.Done()
						return context.Canceled
					},
				)
				mocks = []*mok.Mock{listener.Mock}
			)
			server, errs := MakeServerAndServe(conf, listener, delegate, callback)
			wg.Wait()
			if err := server.Close(); err != nil {
				t.Fatal(err)
			}
			testAsyncErr(wantErr, errs, mocks, t)
		})

	t.Run("We should be able to shutdown the server before it receives a conn",
		func(t *testing.T) {
			var (
				wantErr  = ErrShutdown
				conf     = Conf{WorkersCount: 1}
				listener = func() mock.Listener {
					done := make(chan struct{})
					listener := mock.NewListener().RegisterAccept(
						func() (conn net.Conn, err error) {
							<-done
							err = errors.New("listener closed")
							return
						},
					).RegisterClose(
						func() (err error) {
							close(done)
							return nil
						},
					)
					return listener
				}()
				delegate = mock.NewServerDelegate()
				mocks    = []*mok.Mock{listener.Mock, delegate.Mock}
			)
			server, errs := MakeServerAndServe(conf, listener, delegate, nil)
			time.Sleep(100 * time.Millisecond)
			err := server.Shutdown()
			if err != nil {
				t.Fatal(err)
			}
			testAsyncErr(wantErr, errs, mocks, t)
		})

	t.Run("We should be able to close the server before it receives a conn",
		func(t *testing.T) {
			var (
				wantErr  = ErrClosed
				conf     = Conf{WorkersCount: 1}
				listener = func() mock.Listener {
					done := make(chan struct{})
					listener := mock.NewListener().RegisterAccept(
						func() (conn net.Conn, err error) {
							<-done
							err = errors.New("listener closed")
							return
						},
					).RegisterClose(
						func() (err error) {
							close(done)
							return nil
						},
					)
					return listener
				}()
				delegate = mock.NewServerDelegate()
				mocks    = []*mok.Mock{listener.Mock, delegate.Mock}
			)
			server, errs := MakeServerAndServe(conf, listener, delegate, nil)
			time.Sleep(100 * time.Millisecond)
			err := server.Close()
			if err != nil {
				t.Fatal(err)
			}
			testAsyncErr(wantErr, errs, mocks, t)
		})

	t.Run("Shutdown should fail with an error, if server is not serving",
		func(t *testing.T) {
			var (
				wantErr = ErrNotServing
				conf    = Conf{WorkersCount: 1}
				server  = &Server{Conf: conf}
				err     = server.Shutdown()
			)
			if err != wantErr {
				t.Errorf("unexpected err, want '%v' actual '%v'", wantErr, err)
			}
		})

	t.Run("Close should fail with an error, if server is not serving",
		func(t *testing.T) {
			var (
				wantErr = ErrNotServing
				conf    = Conf{WorkersCount: 1}
				server  = &Server{Conf: conf}
				err     = server.Close()
			)
			if err != wantErr {
				t.Errorf("unexpected err, want '%v' actual '%v'", wantErr, err)
			}
		})

}

func MakeServerAndServe(conf Conf, listener base.Listener,
	delegate mock.ServerDelegate,
	callback LostConnCallback,
) (server *Server, errs <-chan error) {
	server = &Server{Conf: conf, Delegate: delegate, Callback: callback}
	ch := make(chan error, 1)
	go func() {
		err := server.Serve(listener)
		ch <- err
		close(ch)
	}()
	return server, ch
}

func MakeConn(addr net.Addr) mock.Conn {
	return mock.NewConn().RegisterRemoteAddr(
		func() net.Addr { return addr },
	)
}

func SameTime(t1, t2 time.Time) bool {
	return !(t1.Before(t2.Truncate(Delta)) || t1.After(t2.Add(Delta)))
}

func testAsyncErr(wantErr error, errs <-chan error, mocks []*mok.Mock,
	t *testing.T) {
	select {
	case <-time.NewTimer(500 * time.Millisecond).C:
		t.Error("test lasts too long")
	case err := <-errs:
		if err != wantErr {
			t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
		}
	}
	if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
		t.Error(infomap)
	}
}
