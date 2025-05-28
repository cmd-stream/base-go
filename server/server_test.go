package bser

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/cmd-stream/base-go"
	bsmock "github.com/cmd-stream/base-go/server/testdata/mock"
	"github.com/cmd-stream/base-go/testdata/mock"
	asserterror "github.com/ymz-ncnk/assert/error"
	assertfatal "github.com/ymz-ncnk/assert/fatal"
	"github.com/ymz-ncnk/mok"
)

func TestServer(t *testing.T) {

	t.Run("Serving should fail if Conf.WorkersCount == 0", func(t *testing.T) {
		var (
			wantErr = ErrNoWorkers
			server  = &Server{}
			err     = server.Serve(nil)
		)
		asserterror.EqualError(err, wantErr, t)
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
				conn1    = makeConn(addr1)
				addr2    = &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9900}
				conn2    = makeConn(addr2)
				delegate = bsmock.NewDelegate().RegisterNHandle(2,
					func(ctx context.Context, conn net.Conn) (err error) {
						// TODO
						return wantHandleErr
					},
				)
				callback LostConnCallback = func(addr net.Addr, err error) {
					asserterror.EqualError(err, wantHandleErr, t)
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
				}()
				mocks = []*mok.Mock{conn1.Mock, conn2.Mock, listener.Mock}
			)
			server := New(delegate,
				WithWorkersCount(2),
				WithLostConnCallback(callback),
				WithConnReceiver(
					WithFirstConnTimeout(time.Second),
				),
			)
			errs := startServer(server, listener)

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
				addr = &net.TCPAddr{IP: net.ParseIP("127.0.0.1"),
					Port: 9000}
				conn                      = makeConn(addr)
				callback LostConnCallback = func(a net.Addr, err error) {
					if a != addr {
						t.Errorf("unexpected addr, want '%v' actual '%v'", addr, a)
					}
					asserterror.EqualError(err, wantLostConnErr, t)
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
				delegate = bsmock.NewDelegate().RegisterHandle(
					func(ctx context.Context, conn net.Conn) (err error) {
						wg.Done()
						time.Sleep(100 * time.Millisecond)
						return wantLostConnErr
					},
				)
				mocks = []*mok.Mock{listener.Mock}
			)
			server := New(delegate,
				WithWorkersCount(1),
				WithLostConnCallback(callback),
			)
			errs := startServer(server, listener)

			wg.Wait()
			err := server.Shutdown()
			assertfatal.EqualError(err, nil, t)

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
				addr                      = &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9000}
				conn                      = makeConn(addr)
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
				delegate = bsmock.NewDelegate().RegisterHandle(
					func(ctx context.Context, conn net.Conn) (err error) {
						wg.Done()
						<-ctx.Done()
						return context.Canceled
					},
				)
				mocks = []*mok.Mock{listener.Mock}
			)
			server := New(delegate,
				WithWorkersCount(1),
				WithLostConnCallback(callback),
			)
			errs := startServer(server, listener)

			wg.Wait()
			err := server.Close()
			assertfatal.EqualError(err, nil, t)

			testAsyncErr(wantErr, errs, mocks, t)
		})

	t.Run("We should be able to shutdown the server before it receives a conn",
		func(t *testing.T) {
			var (
				wantErr  = ErrShutdown
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
				delegate = bsmock.NewDelegate()
				mocks    = []*mok.Mock{listener.Mock, delegate.Mock}
			)
			server := New(delegate, WithWorkersCount(1))
			errs := startServer(server, listener)

			time.Sleep(100 * time.Millisecond)
			err := server.Shutdown()
			assertfatal.EqualError(err, nil, t)

			testAsyncErr(wantErr, errs, mocks, t)
		})

	t.Run("We should be able to close the server before it receives a conn",
		func(t *testing.T) {
			var (
				wantErr  = ErrClosed
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
				delegate = bsmock.NewDelegate()
				mocks    = []*mok.Mock{listener.Mock, delegate.Mock}
			)
			server := New(delegate, WithWorkersCount(1))
			errs := startServer(server, listener)

			time.Sleep(100 * time.Millisecond)
			err := server.Close()
			assertfatal.EqualError(err, nil, t)

			testAsyncErr(wantErr, errs, mocks, t)
		})

	t.Run("Shutdown should fail with an error, if server is not serving",
		func(t *testing.T) {
			var (
				wantErr = ErrNotServing
				server  = New(nil, WithWorkersCount(1))
				err     = server.Shutdown()
			)
			assertfatal.EqualError(err, wantErr, t)
		})

	t.Run("Close should fail with an error, if server is not serving",
		func(t *testing.T) {
			var (
				wantErr = ErrNotServing
				server  = New(nil, WithWorkersCount(1))
				err     = server.Close()
			)
			assertfatal.EqualError(err, wantErr, t)
		})

}

func startServer(server *Server, listener base.Listener) (
	errs <-chan error) {
	ch := make(chan error, 1)
	go func() {
		err := server.Serve(listener)
		ch <- err
		close(ch)
	}()
	return ch
}

func makeConn(addr net.Addr) mock.Conn {
	return mock.NewConn().RegisterRemoteAddr(func() net.Addr { return addr })
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
	asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
}
