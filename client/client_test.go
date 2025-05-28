package bcln

import (
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/cmd-stream/base-go"
	bcmock "github.com/cmd-stream/base-go/client/testdata/mock"
	"github.com/cmd-stream/base-go/testdata/mock"
	asserterror "github.com/ymz-ncnk/assert/error"
	"github.com/ymz-ncnk/mok"
)

func TestClient(t *testing.T) {

	var delta = 100 * time.Millisecond

	t.Run("We should be able to send cmd and receive several Results",
		func(t *testing.T) {
			var (
				wantSeq     base.Seq = 1
				wantCmd              = mock.NewCmd()
				wantSendN   int      = 1
				wantSendErr error    = nil
				wantResult1          = mock.NewResult().RegisterLastOne(
					func() (lastOne bool) { return false },
				)
				wantBytesRead1 int = 2
				wantResult2        = mock.NewResult().RegisterLastOne(
					func() (lastOne bool) { return true },
				)
				wantBytesRead2 int = 3
				receiveDone        = make(chan struct{})
				delegate           = bcmock.NewDelegate().RegisterSend(
					func(seq base.Seq, cmd base.Cmd[any]) (n int, err error) {
						asserterror.Equal(seq, wantSeq, t)
						asserterror.EqualDeep(cmd, wantCmd, t)
						return wantSendN, nil
					},
				).RegisterFlush(
					func() (err error) { return nil },
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, n int, err error) {
						<-receiveDone
						return wantSeq, wantResult1, wantBytesRead1, nil
					},
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, n int, err error) {
						return wantSeq, wantResult2, wantBytesRead2, nil
					},
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, n int, err error) {
						err = errors.New("Delegate.Receive error")
						return
					},
				).RegisterClose(
					func() (err error) { return nil },
				)
				mocks = []*mok.Mock{wantCmd.Mock, wantResult1.Mock, wantResult2.Mock,
					delegate.Mock}
				client  = New[any](delegate, nil)
				results = make(chan base.AsyncResult, 2)
			)
			seq, n, err := client.Send(wantCmd, results)
			asserterror.EqualError(err, wantSendErr, t)
			asserterror.Equal(seq, wantSeq, t)
			asserterror.Equal(n, wantSendN, t)
			asserterror.Equal(client.Has(seq), true, t)
			close(receiveDone)
			waitDone(client.Done(), t)

			result1 := <-results
			asserterror.EqualDeep(result1.Result, wantResult1, t)
			asserterror.Equal(result1.BytesRead, wantBytesRead1, t)

			result2 := <-results
			asserterror.EqualDeep(result2.Result, wantResult2, t)
			asserterror.Equal(result2.BytesRead, wantBytesRead2, t)

			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

	t.Run("When the client receives the last one Result it should forget cmd",
		func(t *testing.T) {
			var (
				done        = make(chan struct{})
				wantCmd     = mock.NewCmd()
				wantResult1 = mock.NewResult().RegisterLastOne(
					func() (lastOne bool) { return true },
				)
				wantResult1N int = 1
				wantResult2      = mock.NewResult().RegisterLastOne(
					func() (lastOne bool) { return true },
				)
				wantBytesRead2 int = 2
				receiveDone        = make(chan struct{})
				delegate           = bcmock.NewDelegate().RegisterSend(
					func(seq base.Seq, cmd base.Cmd[any]) (n int, err error) {
						return
					},
				).RegisterFlush(
					func() (err error) { return nil },
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, n int, err error) {
						<-receiveDone
						return 1, wantResult1, wantResult1N, nil
					},
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, n int, err error) {
						return 1, wantResult2, wantBytesRead2, nil
					},
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, n int, err error) {
						err = errors.New("Delegate.Receive error")
						return
					},
				).RegisterClose(
					func() (err error) { return nil },
				)
				mocks = []*mok.Mock{wantCmd.Mock, wantResult1.Mock, wantResult2.Mock,
					delegate.Mock}
				callback = func(seq base.Seq, result base.Result) {
					asserterror.EqualDeep(result, wantResult2, t)
					close(done)
				}
				client  = New[any](delegate, WithUnexpectedResultCallback(callback))
				results = make(chan base.AsyncResult, 1)
			)
			client.Send(wantCmd, results)

			close(receiveDone)
			waitDone(done, t)
			waitDone(client.Done(), t)
			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

	t.Run("Send should increment seq", func(t *testing.T) {
		var (
			wantSeq1    base.Seq = 1
			wantSeq2    base.Seq = 2
			wantCmd1             = mock.NewCmd()
			wantCmd2             = mock.NewCmd()
			receiveDone          = make(chan struct{})
			delegate             = bcmock.NewDelegate().RegisterSend(
				func(seq base.Seq, cmd base.Cmd[any]) (n int, err error) {
					asserterror.Equal(seq, wantSeq1, t)
					asserterror.EqualDeep(cmd, wantCmd1, t)
					return
				},
			).RegisterFlush(
				func() (err error) { return nil },
			).RegisterSend(
				func(seq base.Seq, cmd base.Cmd[any]) (n int, err error) {
					asserterror.Equal(seq, wantSeq2, t)
					asserterror.EqualDeep(cmd, wantCmd2, t)
					return
				},
			).RegisterReceive(
				func() (seq base.Seq, result base.Result, n int, err error) {
					<-receiveDone
					err = errors.New("Delegate.Receive error")
					return
				},
			).RegisterFlush(
				func() (err error) { return nil },
			).RegisterClose(
				func() (err error) { return nil },
			)
			mocks  = []*mok.Mock{wantCmd1.Mock, wantCmd2.Mock, delegate.Mock}
			client = New[any](delegate, nil)
		)
		seq, _, _ := client.Send(wantCmd1, nil)
		asserterror.Equal(seq, wantSeq1, t)

		seq, _, _ = client.Send(wantCmd2, nil)
		asserterror.Equal(seq, wantSeq2, t)

		close(receiveDone)
		waitDone(client.Done(), t)
		asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
	})

	t.Run("Send should memorize cmd", func(t *testing.T) {
		var (
			receiveDone = make(chan struct{})
			delegate    = bcmock.NewDelegate().RegisterSend(
				func(seq base.Seq, cmd base.Cmd[any]) (n int, err error) {
					return
				},
			).RegisterFlush(
				func() (err error) { return nil },
			).RegisterReceive(
				func() (seq base.Seq, result base.Result, n int, err error) {
					<-receiveDone
					err = errors.New("Delegate.Receive error")
					return
				},
			).RegisterClose(
				func() (err error) { return nil },
			)
			mocks  = []*mok.Mock{delegate.Mock}
			client = New[any](delegate, nil)
		)
		seq, _, _ := client.Send(nil, nil)
		asserterror.Equal(client.Has(seq), true, t)

		close(receiveDone)
		waitDone(client.Done(), t)
		asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
	})

	t.Run("Seq should be incremented even if Send fails", func(t *testing.T) {
		var (
			wantSeq  base.Seq = 1
			wantErr           = errors.New("Delegate.Send error")
			delegate          = bcmock.NewDelegate().RegisterSend(
				func(seq base.Seq, cmd base.Cmd[any]) (n int, err error) {
					return 1, wantErr
				},
			).RegisterReceive(
				func() (seq base.Seq, result base.Result, n int, err error) {
					err = errors.New("Delegate.Receive error")
					return
				},
			).RegisterClose(
				func() (err error) { return nil },
			)
			mocks  = []*mok.Mock{delegate.Mock}
			client = New[any](delegate, nil)
		)
		seq, _, err := client.Send(nil, nil)
		asserterror.EqualError(err, wantErr, t)
		asserterror.Equal(seq, wantSeq, t)

		waitDone(client.Done(), t)
		asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
	})

	t.Run("If Send fails cmd should be forgotten", func(t *testing.T) {
		var (
			receiveDone = make(chan struct{})
			delegate    = bcmock.NewDelegate().RegisterSend(
				func(seq base.Seq, cmd base.Cmd[any]) (n int, err error) {
					return 0, errors.New("Delegate.Send error")
				},
			).RegisterReceive(
				func() (seq base.Seq, result base.Result, n int, err error) {
					<-receiveDone
					err = errors.New("Delegate.Receive error")
					return
				},
			).RegisterClose(
				func() (err error) { return nil },
			)
			mocks  = []*mok.Mock{delegate.Mock}
			client = New[any](delegate, nil)
		)
		seq, _, _ := client.Send(nil, nil)
		asserterror.Equal(client.Has(seq), false, t)

		close(receiveDone)
		waitDone(client.Done(), t)
		asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
	})

	t.Run("We should be able to send cmd by SendWithDeadline", func(t *testing.T) {
		var (
			wantSeq      base.Seq = 1
			wantDeadline          = time.Now()
			wantCmd               = mock.NewCmd()
			receiveDone           = make(chan struct{})
			delegate              = bcmock.NewDelegate().RegisterSetSendDeadline(
				func(deadline time.Time) (err error) {
					asserterror.SameTime(deadline, wantDeadline, delta, t)
					return nil
				},
			).RegisterSend(
				func(seq base.Seq, cmd base.Cmd[any]) (n int, err error) {
					asserterror.Equal(seq, wantSeq, t)
					asserterror.EqualDeep(cmd, wantCmd, t)
					return
				},
			).RegisterFlush(
				func() (err error) { return nil },
			).RegisterReceive(
				func() (seq base.Seq, result base.Result, n int, err error) {
					<-receiveDone
					err = errors.New("Delegate.Receive error")
					return
				},
			).RegisterClose(
				func() (err error) { return nil },
			)
			mocks  = []*mok.Mock{wantCmd.Mock, delegate.Mock}
			client = New[any](delegate, nil)
		)
		seq, _, err := client.SendWithDeadline(wantCmd, nil, wantDeadline)
		asserterror.EqualError(err, nil, t)
		asserterror.Equal(seq, wantSeq, t)
		asserterror.Equal(client.Has(seq), true, t)

		close(receiveDone)
		waitDone(client.Done(), t)
		asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
	})

	t.Run("Seq should be incremented even if SendWithDeadline fails",
		func(t *testing.T) {
			var (
				wantSeq  base.Seq = 1
				wantErr           = errors.New("Delegate.Send error")
				delegate          = bcmock.NewDelegate().RegisterSetSendDeadline(
					func(deadline time.Time) (err error) {
						return wantErr
					},
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, n int, err error) {
						err = errors.New("Delegate.Receive error")
						return
					},
				).RegisterClose(
					func() (err error) { return nil },
				)
				mocks  = []*mok.Mock{delegate.Mock}
				client = New[any](delegate, nil)
			)
			seq, _, err := client.SendWithDeadline(nil, nil, time.Time{})
			asserterror.EqualError(err, wantErr, t)
			asserterror.Equal(seq, wantSeq, t)

			waitDone(client.Done(), t)
			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

	t.Run("If Delegate.SetSendDeadline fails with an error, SendWithDeadline should return it",
		func(t *testing.T) {
			var (
				wantErr  = errors.New("Delegate.SetSendDeadline error")
				delegate = bcmock.NewDelegate().RegisterSetSendDeadline(
					func(deadline time.Time) (err error) {
						return wantErr
					},
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, n int, err error) {
						err = errors.New("Delegate.Receive error")
						return
					},
				).RegisterClose(
					func() (err error) { return nil },
				)
				mocks  = []*mok.Mock{delegate.Mock}
				client = New[any](delegate, nil)
			)
			_, _, err := client.SendWithDeadline(nil, nil, time.Time{})
			asserterror.EqualError(err, wantErr, t)
			waitDone(client.Done(), t)
			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

	t.Run("If Delegate.Send fails with an error, SendWithDeadline should return it",
		func(t *testing.T) {
			var (
				wantErr  = errors.New("Delegate.Send error")
				delegate = bcmock.NewDelegate().RegisterSetSendDeadline(
					func(deadline time.Time) (err error) {
						return nil
					},
				).RegisterSend(
					func(seq base.Seq, cmd base.Cmd[any]) (n int, err error) {
						return 0, wantErr
					},
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, n int, err error) {
						err = errors.New("Delegate.Receive error")
						return
					},
				).RegisterClose(
					func() (err error) { return nil },
				)
				mocks  = []*mok.Mock{delegate.Mock}
				client = New[any](delegate, nil)
			)
			_, _, err := client.SendWithDeadline(nil, nil, time.Time{})
			asserterror.EqualError(err, wantErr, t)

			waitDone(client.Done(), t)
			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

	t.Run("Client should forget cmd, if SendWithDeadline failed, because of Delegate.SetSendDeadline",
		func(t *testing.T) {
			var (
				delegate = bcmock.NewDelegate().RegisterSetSendDeadline(
					func(deadline time.Time) (err error) {
						return errors.New("Delegate.SetSendDeadline error")
					},
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, n int, err error) {
						err = errors.New("Delegate.Receive error")
						return
					},
				).RegisterClose(
					func() (err error) { return nil },
				)
				mocks  = []*mok.Mock{delegate.Mock}
				client = New[any](delegate, nil)
			)
			seq, _, _ := client.SendWithDeadline(nil, nil, time.Time{})
			asserterror.Equal(client.Has(seq), false, t)

			waitDone(client.Done(), t)
			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

	t.Run("Client should forget cmd, if SendWithDeadline failed, because of Delegate.Send",
		func(t *testing.T) {
			var (
				delegate = bcmock.NewDelegate().RegisterSetSendDeadline(
					func(deadline time.Time) (err error) {
						return nil
					},
				).RegisterSend(
					func(seq base.Seq, cmd base.Cmd[any]) (n int, err error) {
						err = errors.New("Delegate.Send error")
						return
					},
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, n int, err error) {
						err = errors.New("Delegate.Receive error")
						return
					},
				).RegisterClose(
					func() (err error) { return nil },
				)
				mocks  = []*mok.Mock{delegate.Mock}
				client = New[any](delegate, nil)
			)
			seq, _, _ := client.SendWithDeadline(nil, nil, time.Time{})
			asserterror.Equal(client.Has(seq), false, t)

			waitDone(client.Done(), t)
			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

	t.Run("We should be able to forget cmd", func(t *testing.T) {
		var (
			cmd         = mock.NewCmd()
			receiveDone = make(chan struct{})
			delegate    = bcmock.NewDelegate().RegisterSend(
				func(seq base.Seq, cmd base.Cmd[any]) (n int, err error) {
					return
				},
			).RegisterFlush(
				func() (err error) { return nil },
			).RegisterReceive(
				func() (seq base.Seq, result base.Result, n int, err error) {
					<-receiveDone
					err = errors.New("Delegate.Receive error")
					return
				},
			).RegisterClose(
				func() (err error) { return nil },
			)
			mocks  = []*mok.Mock{cmd.Mock, delegate.Mock}
			client = New[any](delegate, nil)
		)
		seq, _, _ := client.Send(cmd, nil)
		client.Forget(seq)
		asserterror.Equal(client.Has(seq), false, t)

		close(receiveDone)
		waitDone(client.Done(), t)
		asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
	})

	t.Run("If the client was closed or failed to receive a next result, Done channel should be closed and Err method should return the cause",
		func(t *testing.T) {
			var (
				wantErr  = errors.New("Delegate.Receive error")
				delegate = bcmock.NewDelegate().RegisterReceive(
					func() (seq base.Seq, result base.Result, n int, err error) {
						err = wantErr
						return
					},
				).RegisterClose(
					func() (err error) { return nil },
				)
				mocks  = []*mok.Mock{delegate.Mock}
				client = New[any](delegate, nil)
			)
			waitDone(client.Done(), t)
			err := client.Err()
			asserterror.EqualError(err, wantErr, t)
			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

	t.Run("We should be able to close the client while it queues a result",
		func(t *testing.T) {
			var (
				wantCmd    = mock.NewCmd()
				wantResult = mock.NewResult().RegisterLastOne(
					func() (lastOne bool) { return true },
				)
				results  = make(chan base.AsyncResult)
				delegate = bcmock.NewDelegate().RegisterSend(
					func(seq base.Seq, cmd base.Cmd[any]) (n int, err error) {
						err = errors.New("Delegate.Send error")
						return
					},
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, n int, err error) {
						return 1, wantResult, 0, nil
					},
				).RegisterClose(
					func() (err error) { return nil },
				)
				mocks  = []*mok.Mock{wantCmd.Mock, wantResult.Mock, delegate.Mock}
				client = New[any](delegate, nil)
			)
			client.Send(wantCmd, results)
			time.Sleep(100 * time.Millisecond)
			client.Close()
			waitDone(client.Done(), t)

			err := client.Err()
			asserterror.EqualError(err, ErrClosed, t)
			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

	t.Run("If Delegate.Close fails with an error, Close should return it",
		func(t *testing.T) {
			var (
				receiveDone = make(chan struct{})
				wantErr     = errors.New("Delegate.Close error")
				delegate    = bcmock.NewDelegate().RegisterReceive(
					func() (seq base.Seq, result base.Result, n int, err error) {
						<-receiveDone
						err = errors.New("Delegate.Receive error")
						return
					},
				).RegisterClose(
					func() (err error) {
						return wantErr
					},
				).RegisterClose(
					func() (err error) { return nil },
				)
				mocks  = []*mok.Mock{delegate.Mock}
				client = New[any](delegate, nil)
			)
			err := client.Close()
			asserterror.EqualError(err, wantErr, t)

			close(receiveDone)
			waitDone(client.Done(), t)
			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

	t.Run("If Delegate.Flush fails with an error, Send of all involved Commands should return error",
		func(t *testing.T) {
			var (
				wantErr  = errors.New("flush error")
				cmd1     = mock.NewCmd()
				cmd2     = mock.NewCmd()
				cmd3     = mock.NewCmd()
				delegate = bcmock.NewDelegate().RegisterNSend(3,
					func(seq base.Seq, cmd base.Cmd[any]) (n int, err error) {
						return
					},
				).RegisterNFlush(3,
					func() (err error) { return wantErr },
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, n int, err error) {
						err = errors.New("Delegate.Receive error")
						return
					},
				).RegisterClose(
					func() (err error) { return nil },
				)
				wg = &sync.WaitGroup{}
				// mocks  = []*mok.Mock{delegate.Mock}
				client = New[any](delegate, nil)
			)
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _, err := client.Send(cmd1, nil)
				if err != wantErr {
					t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
				}
			}()
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _, err := client.Send(cmd2, nil)
				asserterror.EqualError(err, wantErr, t)
			}()
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _, err := client.Send(cmd3, nil)
				asserterror.EqualError(err, wantErr, t)
			}()
			wg.Wait()
			waitDone(client.Done(), t)
			// Commented out, because we do not know the actual count of the
			// Delegate.Flush() method calls.
			//
			// If we want the flush method to be called only once, we should put
			// time.Sleep(200*time.Milisecond) before the flush method acquires a
			// lock.
			//
			// asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

	t.Run("If the client has lost a connection it should try to reconnect",
		func(t *testing.T) {
			var (
				reconected = make(chan struct{})
				delegate   = bcmock.NewReconnectDelegate().RegisterReceive(
					func() (seq base.Seq, result base.Result, n int, err error) {
						err = net.ErrClosed
						return
					},
				).RegisterReconnect(
					func() error { close(reconected); return nil },
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, n int, err error) {
						err = errors.New("Delegate.Receive error")
						return
					},
				).RegisterClose(
					func() (err error) { return nil },
				)
				mocks  = []*mok.Mock{delegate.Mock}
				client = New[any](delegate, nil)
			)
			waitDone(reconected, t)
			waitDone(client.Done(), t)
			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

	t.Run("If the client is closed it should not reconnect", func(t *testing.T) {
		var (
			receiveDone = make(chan struct{})
			delegate    = bcmock.NewReconnectDelegate().RegisterReceive(
				func() (seq base.Seq, result base.Result, n int, err error) {
					<-receiveDone
					err = ErrClosed
					return
				},
			).RegisterClose(
				func() (err error) { close(receiveDone); return nil },
			)
			mocks  = []*mok.Mock{delegate.Mock}
			client = New[any](delegate, nil)
		)
		err := client.Close()
		asserterror.EqualError(err, nil, t)

		waitDone(client.Done(), t)
		asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
	})

	t.Run("If reconnection fails with an error, it should became the client error",
		func(t *testing.T) {
			var (
				wantErr  = errors.New("reconnection error")
				delegate = bcmock.NewReconnectDelegate().RegisterReceive(
					func() (seq base.Seq, result base.Result, n int, err error) {
						err = net.ErrClosed
						return
					},
				).RegisterReconnect(
					func() error { return wantErr },
				).RegisterClose(
					func() (err error) { return nil },
				)
				mocks  = []*mok.Mock{delegate.Mock}
				client = New[any](delegate, nil)
			)
			waitDone(client.Done(), t)
			err := client.Err()
			asserterror.EqualError(err, wantErr, t)

			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

	t.Run("Upon creation, the client should call KeepaliveDelegate.Keepalive()",
		func(t *testing.T) {
			var (
				delegate = bcmock.NewKeepaliveDelegate().RegisterKeepalive(
					func(muSn *sync.Mutex) {},
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, n int, err error) {
						err = errors.New("Delegate.Receive error")
						return
					},
				).RegisterClose(
					func() (err error) { return },
				)
				client = New[any](delegate, nil)
				mocks  = []*mok.Mock{delegate.Mock}
			)
			waitDone(client.Done(), t)
			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

}

func waitDone(done <-chan struct{}, t *testing.T) {
	select {
	case <-done:
	case <-time.NewTimer(time.Second).C:
		t.Fatal("test lasts too long")
	}
}
