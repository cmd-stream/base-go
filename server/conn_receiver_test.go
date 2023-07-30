package server

import (
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/cmd-stream/base-go/testdata/mock"
	"github.com/ymz-ncnk/mok"
)

func TestConnReceiver(t *testing.T) {

	t.Run("Conf.FirstConnTimeout should be applied only to the first conn + if Listener.SetDeadline failed with an error, Run should return it",
		func(t *testing.T) {
			var (
				wantErr   = errors.New("SetDeadline error")
				startTime = time.Now()
				conf      = ConnReceiverConf{FirstConnTimeout: time.Second}
				listener  = mock.NewListener().RegisterSetDeadline(
					func(deadline time.Time) (err error) {
						wantDeadline := startTime.Add(conf.FirstConnTimeout)
						if !SameTime(deadline, wantDeadline) {
							return fmt.Errorf("unexpected deadline, want '%v' actual '%v'",
								wantDeadline,
								deadline)
						}
						return wantErr
					},
				)
				mocks    = []*mok.Mock{listener.Mock}
				receiver = NewConnReceiver(conf, listener, make(chan net.Conn))
			)
			testConnReceiver(receiver, wantErr, mocks, t)
		})

	t.Run("If accepting of the first conn failed with an error, Run should return it",
		func(t *testing.T) {
			var (
				wantErr  = errors.New("accept error")
				listener = mock.NewListener().RegisterAccept(
					func() (net.Conn, error) { return nil, wantErr },
				)
				conns    = make(chan net.Conn)
				mocks    = []*mok.Mock{listener.Mock}
				receiver = NewConnReceiver(ConnReceiverConf{}, listener, conns)
			)
			testConnReceiver(receiver, wantErr, mocks, t)
		})

	t.Run("Conf.FirstConnTimeout should be applied to only first conn + if cancelation of the first conn deadline failed with an error, Run should return it",
		func(t *testing.T) {
			var (
				wantErr  = errors.New("set deadline error")
				wantConn = mock.NewConn().RegisterClose(
					func() (err error) { return nil },
				)
				startTime = time.Now()
				conf      = ConnReceiverConf{FirstConnTimeout: time.Second}
				listener  = mock.NewListener().RegisterSetDeadline(
					func(deadline time.Time) (err error) {
						wantDeadline := startTime.Add(conf.FirstConnTimeout)
						if !SameTime(deadline, wantDeadline) {
							return fmt.Errorf("unexpected deadline, want '%v' actual '%v'",
								wantDeadline,
								deadline)
						}
						return
					},
				).RegisterAccept(
					func() (conn net.Conn, err error) {
						return wantConn, nil
					},
				).RegisterSetDeadline(
					func(deadline time.Time) (err error) {
						if !deadline.IsZero() {
							return fmt.Errorf("unexpected deadline, want '%v' actual '%v'",
								time.Time{},
								deadline)
						}
						return wantErr
					},
				)
				mocks    = []*mok.Mock{wantConn.Mock, listener.Mock}
				receiver = NewConnReceiver(conf, listener, make(chan net.Conn, 1))
			)
			testConnReceiver(receiver, wantErr, mocks, t)
		})

	t.Run("If Listener.Accept for the first conn failed with an error, Run should return it",
		func(t *testing.T) {
			var (
				wantErr = errors.New("set deadline error")

				listener = mock.NewListener().RegisterAccept(
					func() (conn net.Conn, err error) {
						return nil, wantErr
					},
				)
				mocks    = []*mok.Mock{listener.Mock}
				receiver = NewConnReceiver(ConnReceiverConf{}, listener, make(chan net.Conn, 1))
			)
			testConnReceiver(receiver, wantErr, mocks, t)
		})

	t.Run("If Listener.Accept for the second conn failed with an error, Run should return it",
		func(t *testing.T) {
			var (
				wantErr  = errors.New("set deadline error")
				wantConn = mock.NewConn().RegisterClose(
					func() (err error) { return nil },
				)
				listener = mock.NewListener().RegisterAccept(
					func() (conn net.Conn, err error) {
						return wantConn, nil
					},
				).RegisterAccept(
					func() (conn net.Conn, err error) {
						return nil, wantErr
					},
				)
				mocks    = []*mok.Mock{wantConn.Mock, listener.Mock}
				receiver = NewConnReceiver(ConnReceiverConf{}, listener, make(chan net.Conn, 1))
			)
			testConnReceiver(receiver, wantErr, mocks, t)
		})

	t.Run("ConnReceiver should be able to accept several connections",
		func(t *testing.T) {
			var (
				// wantErr  = errors.New("done")
				done     = make(chan struct{})
				conn1    = mock.NewConn()
				conn2    = mock.NewConn()
				listener = mock.NewListener().RegisterAccept(
					func() (net.Conn, error) { return conn1, nil },
				).RegisterAccept(
					func() (net.Conn, error) { return conn2, nil },
				).RegisterAccept(
					func() (net.Conn, error) {
						<-done
						return nil, errors.New("done")
					},
				).RegisterClose(
					func() error { close(done); return nil },
				)
				conns    = make(chan net.Conn, 2)
				mocks    = []*mok.Mock{conn1.Mock, conn2.Mock, listener.Mock}
				receiver = NewConnReceiver(ConnReceiverConf{}, listener, conns)
			)
			go func() {
				i := 0
				for {
					select {
					case <-conns:
						i++
						if i == 2 {
							goto Stop
						}
					case <-time.NewTimer(200 * time.Millisecond).C:
						panic("test lasts too long")
					}
				}
			Stop:
				if err := receiver.Stop(); err != nil {
					t.Error(err)
				}
			}()
			testConnReceiver(receiver, ErrClosed, mocks, t)
		})

	t.Run("We should be able to close the ConnHandler while Listener.Accept",
		func(t *testing.T) {
			testStopWhileAccept(false, t)
		})

	t.Run("We should be able to shutdown the ConnHandler while Listener.Accept",
		func(t *testing.T) {
			testStopWhileAccept(true, t)
		})

	t.Run("We should be able to close the ConnHandler while it adds conn to queue",
		func(t *testing.T) {
			testStopWhileQueueConn(false, t)
		})

	t.Run("We should be able to shutdown the ConnHandler while it adds conn to queue",
		func(t *testing.T) {
			testStopWhileQueueConn(true, t)
		})

}

func RunConnReceiver(r *ConnReceiver) (errs chan error) {
	errs = make(chan error, 1)
	go func() {
		if err := r.Run(); err != nil {
			errs <- err
		}
		close(errs)
	}()
	return
}

func testConnReceiver(r *ConnReceiver, wantErr error, mocks []*mok.Mock,
	t *testing.T) {
	errs := RunConnReceiver(r)
	testAsyncErr(wantErr, errs, mocks, t)
}

func testStopWhileAccept(shutdown bool, t *testing.T) {
	// wantErr := errors.New("accept failed, cause close")
	wantErr := ErrClosed
	if shutdown {
		wantErr = nil
	}
	var (
		listener = func() mock.Listener {
			done := make(chan error)
			return mock.NewListener().RegisterAccept(
				func() (net.Conn, error) {
					<-done
					if shutdown {
						return nil, errors.New("accept failed, cause shutdown")
					}
					return nil, wantErr
				},
			).RegisterClose(
				func() error { close(done); return nil },
			)
		}()
		conns = make(chan net.Conn)
		mocks = []*mok.Mock{listener.Mock}
	)
	receiver := NewConnReceiver(ConnReceiverConf{}, listener, conns)
	go func() {
		time.Sleep(100 * time.Millisecond)
		if shutdown {
			receiver.Shutdown()
		} else {
			receiver.Stop()
		}
	}()
	testConnReceiver(receiver, wantErr, mocks, t)
	_, more := <-conns
	if more {
		t.Error("conns chan is not closed")
	}
}

func testStopWhileQueueConn(shutdown bool, t *testing.T) {
	wantErr := ErrClosed
	if shutdown {
		wantErr = nil
	}
	var (
		done = make(chan error)
		conn = func() mock.Conn {
			conn := mock.NewConn().RegisterClose(
				func() (err error) { return nil },
			)
			return conn
		}()
		listener = mock.NewListener().RegisterAccept(
			func() (net.Conn, error) {
				return conn, nil
			},
		).RegisterClose(
			func() error { close(done); return nil },
		)
		conns    = make(chan net.Conn)
		mocks    = []*mok.Mock{conn.Mock, listener.Mock}
		receiver = NewConnReceiver(ConnReceiverConf{}, listener, conns)
	)
	go func() {
		time.Sleep(100 * time.Millisecond)
		if shutdown {
			receiver.Shutdown()
		} else {
			receiver.Stop()
		}
	}()
	testConnReceiver(receiver, wantErr, mocks, t)
	_, more := <-conns
	if more {
		t.Error("conns chan is not closed")
	}
}
