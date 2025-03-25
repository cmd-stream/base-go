package bser

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"

	"github.com/cmd-stream/base-go/testdata/mock"
	"github.com/ymz-ncnk/mok"
)

func TestWorker(t *testing.T) {

	addr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9000}

	t.Run("Worker should be able to handle several conns with LostConnCallback",
		func(t *testing.T) {
			var (
				wantErr  = errors.New("handle conn failed")
				wg       = &sync.WaitGroup{}
				conn1    = MakeConn(addr)
				conn2    = MakeConn(addr)
				delegate = mock.NewServerDelegate().RegisterHandle(
					func(ctx context.Context, conn net.Conn) (err error) {
						if conn != conn1 {
							t.Errorf("unexpected conn, want '%v' actual '%v'", conn1, conn)
						}
						return wantErr
					},
				).RegisterHandle(
					func(ctx context.Context, conn net.Conn) (err error) {
						if conn != conn2 {
							t.Errorf("unexpected conn, want '%v' actual '%v'", conn1, conn)
						}
						return wantErr
					},
				)
				lostConnCallback = func(addr net.Addr, err error) {
					defer wg.Done()
					if err != wantErr {
						t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
					}
				}
			)
			testWorker(delegate, conn1, conn2, lostConnCallback, wg, t)
		})

	t.Run("Worker should be able to handle several conns without LostConnCallback",
		func(t *testing.T) {
			var (
				wantErr  = errors.New("handle conn failed")
				wg       = &sync.WaitGroup{}
				conn1    = mock.NewConn()
				conn2    = mock.NewConn()
				delegate = mock.NewServerDelegate().RegisterHandle(
					func(ctx context.Context, conn net.Conn) (err error) {
						defer wg.Done()
						if conn != conn1 {
							t.Errorf("unexpected conn, want '%v' actual '%v'", conn1, conn)
						}
						return wantErr
					},
				).RegisterHandle(
					func(ctx context.Context, conn net.Conn) (err error) {
						defer wg.Done()
						if conn != conn2 {
							t.Errorf("unexpected conn, want '%v' actual '%v'", conn1, conn)
						}
						return wantErr
					},
				)
			)
			testWorker(delegate, conn1, conn2, nil, wg, t)
		},
	)

	// t.Run("Handle two connections", func(t *testing.T) {
	// 	// addr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9000}

	// 	// t.Run("With LostConnCallback", func(t *testing.T) {
	// 	// 	var (
	// 	// 		wantErr  = errors.New("handle conn failed")
	// 	// 		wg       = &sync.WaitGroup{}
	// 	// 		conn1    = MakeConn(addr)
	// 	// 		conn2    = MakeConn(addr)
	// 	// 		delegate = mock.NewServerDelegate().RegisterHandle(
	// 	// 			func(ctx context.Context, conn net.Conn) (err error) {
	// 	// 				if conn != conn1 {
	// 	// 					t.Errorf("unexpected conn, want '%v' actual '%v'", conn1, conn)
	// 	// 				}
	// 	// 				return wantErr
	// 	// 			},
	// 	// 		).RegisterHandle(
	// 	// 			func(ctx context.Context, conn net.Conn) (err error) {
	// 	// 				if conn != conn2 {
	// 	// 					t.Errorf("unexpected conn, want '%v' actual '%v'", conn1, conn)
	// 	// 				}
	// 	// 				return wantErr
	// 	// 			},
	// 	// 		)
	// 	// 		lostConnCallback = func(addr net.Addr, err error) {
	// 	// 			defer wg.Done()
	// 	// 			if err != wantErr {
	// 	// 				t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
	// 	// 			}
	// 	// 		}
	// 	// 	)
	// 	// 	testWorker(delegate, conn1, conn2, lostConnCallback, wg, t)
	// 	// })

	// 	// t.Run("Without LostConnCallback", func(t *testing.T) {
	// 	// 	var (
	// 	// 		wantErr  = errors.New("handle conn failed")
	// 	// 		wg       = &sync.WaitGroup{}
	// 	// 		conn1    = mock.NewConn()
	// 	// 		conn2    = mock.NewConn()
	// 	// 		delegate = mock.NewServerDelegate().RegisterHandle(
	// 	// 			func(ctx context.Context, conn net.Conn) (err error) {
	// 	// 				defer wg.Done()
	// 	// 				if conn != conn1 {
	// 	// 					t.Errorf("unexpected conn, want '%v' actual '%v'", conn1, conn)
	// 	// 				}
	// 	// 				return wantErr
	// 	// 			},
	// 	// 		).RegisterHandle(
	// 	// 			func(ctx context.Context, conn net.Conn) (err error) {
	// 	// 				defer wg.Done()
	// 	// 				if conn != conn2 {
	// 	// 					t.Errorf("unexpected conn, want '%v' actual '%v'", conn1, conn)
	// 	// 				}
	// 	// 				return wantErr
	// 	// 			},
	// 	// 		)
	// 	// 	)
	// 	// 	testWorker(delegate, conn1, conn2, nil, wg, t)
	// 	// })
	// })

	t.Run("We should be able to close the worker", func(t *testing.T) {
		var (
			wantErr  = ErrClosed
			conns    = make(chan net.Conn)
			delegate = mock.NewServerDelegate()
			mocks    = []*mok.Mock{delegate.Mock}
			worker   = NewWorker(conns, delegate, nil)
		)
		errs := RunWorker(worker)
		if err := worker.Stop(); err != nil {
			t.Fatal(err)
		}
		testAsyncErr(wantErr, errs, mocks, t)
	})

	t.Run("We should be able to shutdown the worker", func(t *testing.T) {
		var (
			wantErr  error = nil
			conns          = make(chan net.Conn)
			delegate       = mock.NewServerDelegate()
			mocks          = []*mok.Mock{delegate.Mock}
			worker         = NewWorker(conns, delegate, nil)
		)
		errs := RunWorker(worker)
		close(conns)
		testAsyncErr(wantErr, errs, mocks, t)
	})

}

func RunWorker(worker Worker) (errs chan error) {
	errs = make(chan error, 1)
	go func() {
		err := worker.Run()
		errs <- err
		close(errs)
	}()
	return errs
}

func testWorker(delegate mock.ServerDelegate, conn1, conn2 mock.Conn,
	lostConnCallback LostConnCallback,
	wg *sync.WaitGroup,
	t *testing.T,
) {
	var (
		conns  = make(chan net.Conn)
		worker = NewWorker(conns, delegate, lostConnCallback)
		mocks  = []*mok.Mock{conn1.Mock, conn2.Mock, delegate.Mock}
		errs   = RunWorker(worker)
	)

	wg.Add(2)
	conns <- conn1
	conns <- conn2
	wg.Wait()

	err := worker.Stop()
	if err != nil {
		t.Fatal(err)
	}
	testAsyncErr(ErrClosed, errs, mocks, t)
}
