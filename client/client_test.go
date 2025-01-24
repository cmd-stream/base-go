package client

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/cmd-stream/base-go"
	"github.com/cmd-stream/base-go/testdata/mock"
	"github.com/ymz-ncnk/mok"
)

const Delta = 100 * time.Millisecond

func TestClient(t *testing.T) {

	t.Run("We should be able to send cmd and receive several Results",
		func(t *testing.T) {
			var (
				wantSeq     base.Seq = 1
				wantCmd              = mock.NewCmd()
				wantResult1          = mock.NewResult().RegisterLastOne(
					func() (lastOne bool) { return false },
				)
				wantResult2 = mock.NewResult().RegisterLastOne(
					func() (lastOne bool) { return true },
				)
				receiveDone = make(chan struct{})
				delegate    = mock.NewClientDelegate().RegisterSend(
					func(seq base.Seq, cmd base.Cmd[any]) (err error) {
						if seq != wantSeq {
							return fmt.Errorf("unexpected seq, want '%v' actual '%v'",
								wantSeq,
								seq)
						}
						if cmd != wantCmd {
							return fmt.Errorf("unexpected cmd, want '%v' actual '%v'",
								wantCmd,
								cmd)
						}
						return
					},
				).RegisterFlush(func() (err error) {
					return nil
				}).RegisterReceive(
					func() (seq base.Seq, result base.Result, err error) {
						<-receiveDone
						return wantSeq, wantResult1, nil
					},
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, err error) {
						return wantSeq, wantResult2, nil
					},
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, err error) {
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
			seq, err := client.Send(wantCmd, results)
			if err != nil {
				t.Errorf("unexpected error, watn '%v' actual %v", nil, err)
			}
			if seq != wantSeq {
				t.Errorf("unexpected seq, want '%v' actual '%v'", wantSeq, seq)
			}
			if !client.Has(seq) {
				t.Error("cmd was not memorized")
			}
			close(receiveDone)
			waitDone(client.Done(), t)
			result1 := <-results
			if result1.Result != wantResult1 {
				t.Errorf("unexpected Result, want '%v' actual '%v'", wantResult1,
					result1)
			}
			result2 := <-results
			if result2.Result != wantResult2 {
				t.Errorf("unexpected Result, want '%v' actual '%v'", wantResult2,
					result2.Result)
			}
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

	t.Run("When the client receives the last one Result it should forget cmd",
		func(t *testing.T) {
			var (
				done        = make(chan struct{})
				wantCmd     = mock.NewCmd()
				wantResult1 = mock.NewResult().RegisterLastOne(
					func() (lastOne bool) { return true },
				)
				wantResult2 = mock.NewResult().RegisterLastOne(
					func() (lastOne bool) { return true },
				)
				receiveDone = make(chan struct{})
				delegate    = mock.NewClientDelegate().RegisterSend(
					func(seq base.Seq, cmd base.Cmd[any]) (err error) {
						return
					},
				).RegisterFlush(
					func() (err error) { return nil },
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, err error) {
						<-receiveDone
						return 1, wantResult1, nil
					},
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, err error) {
						return 1, wantResult2, nil
					},
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, err error) {
						err = errors.New("Delegate.Receive error")
						return
					},
				).RegisterClose(
					func() (err error) { return nil },
				)
				mocks = []*mok.Mock{wantCmd.Mock, wantResult1.Mock, wantResult2.Mock,
					delegate.Mock}
				callback = func(seq base.Seq, result base.Result) {
					if result != wantResult2 {
						t.Errorf("unexpected result, want '%v' actual '%v'", wantResult2,
							result)
					}
					close(done)
				}
				client  = New[any](delegate, callback)
				results = make(chan base.AsyncResult, 1)
			)
			client.Send(wantCmd, results)
			close(receiveDone)
			waitDone(done, t)
			waitDone(client.Done(), t)
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

	t.Run("Send should increment seq", func(t *testing.T) {
		var (
			wantSeq1    base.Seq = 1
			wantSeq2    base.Seq = 2
			wantCmd1             = mock.NewCmd()
			wantCmd2             = mock.NewCmd()
			receiveDone          = make(chan struct{})
			delegate             = mock.NewClientDelegate().RegisterSend(
				func(seq base.Seq, cmd base.Cmd[any]) (err error) {
					if seq != wantSeq1 {
						t.Errorf("unexpected seq, want '%v' actual '%v'", wantSeq1, seq)
					}
					if cmd != wantCmd1 {
						t.Errorf("unexpected cmd, want '%v' actual '%v'", wantCmd1, cmd)
					}
					return
				},
			).RegisterFlush(
				func() (err error) { return nil },
			).RegisterSend(
				func(seq base.Seq, cmd base.Cmd[any]) (err error) {
					if seq != wantSeq2 {
						t.Errorf("unexpected seq, want '%v' actual '%v'", wantSeq1, seq)
					}
					if cmd != wantCmd2 {
						t.Errorf("unexpected cmd, want '%v' actual '%v'", wantCmd2, cmd)
					}
					return
				},
			).RegisterReceive(
				func() (seq base.Seq, result base.Result, err error) {
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
		seq, _ := client.Send(wantCmd1, nil)
		if seq != wantSeq1 {
			t.Errorf("unexpected seq, want '%v' actual '%v'", wantSeq1, seq)
		}
		seq, _ = client.Send(wantCmd2, nil)
		if seq != wantSeq2 {
			t.Errorf("unexpected seq, want '%v' actual '%v'", wantSeq2, seq)
		}
		close(receiveDone)
		waitDone(client.Done(), t)
		if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
			t.Error(infomap)
		}
	})

	t.Run("Send should memorize cmd", func(t *testing.T) {
		var (
			receiveDone = make(chan struct{})
			delegate    = mock.NewClientDelegate().RegisterSend(
				func(seq base.Seq, cmd base.Cmd[any]) (err error) {
					return nil
				},
			).RegisterFlush(
				func() (err error) { return nil },
			).RegisterReceive(
				func() (seq base.Seq, result base.Result, err error) {
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
		seq, _ := client.Send(nil, nil)
		if !client.Has(seq) {
			t.Error("cmd was not memorized")
		}
		close(receiveDone)
		waitDone(client.Done(), t)
		if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
			t.Error(infomap)
		}
	})

	t.Run("Seq should be incremented even if Send fails", func(t *testing.T) {
		var (
			wantSeq  base.Seq = 1
			wantErr           = errors.New("Delegate.Send error")
			delegate          = mock.NewClientDelegate().RegisterSend(
				func(seq base.Seq, cmd base.Cmd[any]) (err error) {
					return wantErr
				},
			).RegisterReceive(
				func() (seq base.Seq, result base.Result, err error) {
					err = errors.New("Delegate.Receive error")
					return
				},
			).RegisterClose(
				func() (err error) { return nil },
			)
			mocks  = []*mok.Mock{delegate.Mock}
			client = New[any](delegate, nil)
		)
		seq, err := client.Send(nil, nil)
		if err != wantErr {
			t.Errorf("unexpected err, want '%v' actual '%v'", wantErr, err)
		}
		if seq != wantSeq {
			t.Errorf("unexpected seq, want '%v' actual '%v'", wantSeq, seq)
		}
		waitDone(client.Done(), t)
		if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
			t.Error(infomap)
		}
	})

	t.Run("If Send fails cmd should be forgotten", func(t *testing.T) {
		var (
			receiveDone = make(chan struct{})
			delegate    = mock.NewClientDelegate().RegisterSend(
				func(seq base.Seq, cmd base.Cmd[any]) (err error) {
					return errors.New("Delegate.Send error")
				},
			).RegisterReceive(
				func() (seq base.Seq, result base.Result, err error) {
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
		seq, _ := client.Send(nil, nil)
		if client.Has(seq) {
			t.Error("cmd was not forgotten")
		}
		close(receiveDone)
		waitDone(client.Done(), t)
		if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
			t.Error(infomap)
		}
	})

	t.Run("We should be able to send cmd by SendWithDeadline", func(t *testing.T) {
		var (
			wantSeq      base.Seq = 1
			wantDeadline          = time.Now()
			wantCmd               = mock.NewCmd()
			receiveDone           = make(chan struct{})
			delegate              = mock.NewClientDelegate().RegisterSetSendDeadline(
				func(deadline time.Time) (err error) {
					if !SameTime(deadline, wantDeadline) {
						return fmt.Errorf("unexpected deadline, want '%v' actual '%v'",
							wantDeadline,
							deadline)
					}
					return nil
				},
			).RegisterSend(
				func(seq base.Seq, cmd base.Cmd[any]) (err error) {
					if seq != wantSeq {
						return fmt.Errorf("unexpected seq, want '%v' actual '%v'", wantSeq,
							seq)
					}
					if cmd != wantCmd {
						return fmt.Errorf("unexpected cmd, want '%v' actual '%v'", wantCmd,
							cmd)
					}
					return
				},
			).RegisterFlush(
				func() (err error) { return nil },
			).RegisterReceive(
				func() (seq base.Seq, result base.Result, err error) {
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
		seq, err := client.SendWithDeadline(wantDeadline, wantCmd, nil)
		if err != nil {
			t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
		}
		if seq != wantSeq {
			t.Errorf("unexpected seq, want '%v' actual '%v'", wantSeq, seq)
		}
		if !client.Has(seq) {
			t.Error("cmd was not memorized")
		}
		close(receiveDone)
		waitDone(client.Done(), t)
		if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
			t.Error(infomap)
		}
	})

	t.Run("Seq should be incremented even if SendWithDeadline fails",
		func(t *testing.T) {
			var (
				wantSeq  base.Seq = 1
				wantErr           = errors.New("Delegate.Send error")
				delegate          = mock.NewClientDelegate().RegisterSetSendDeadline(
					func(deadline time.Time) (err error) {
						return wantErr
					},
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, err error) {
						err = errors.New("Delegate.Receive error")
						return
					},
				).RegisterClose(
					func() (err error) { return nil },
				)
				mocks  = []*mok.Mock{delegate.Mock}
				client = New[any](delegate, nil)
			)
			seq, err := client.SendWithDeadline(time.Time{}, nil, nil)
			if err != wantErr {
				t.Errorf("unexpected err, want '%v' actual '%v'", wantErr, err)
			}
			if seq != wantSeq {
				t.Errorf("unexpected seq, want '%v' actual '%v'", wantSeq, seq)
			}
			waitDone(client.Done(), t)
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

	t.Run("If Delegate.SetSendDeadline fails with an error, SendWithDeadline should return it",
		func(t *testing.T) {
			var (
				wantErr  = errors.New("Delegate.SetSendDeadline error")
				delegate = mock.NewClientDelegate().RegisterSetSendDeadline(
					func(deadline time.Time) (err error) {
						return wantErr
					},
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, err error) {
						err = errors.New("Delegate.Receive error")
						return
					},
				).RegisterClose(
					func() (err error) { return nil },
				)
				mocks  = []*mok.Mock{delegate.Mock}
				client = New[any](delegate, nil)
			)
			_, err := client.SendWithDeadline(time.Time{}, nil, nil)
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
			}
			waitDone(client.Done(), t)
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

	t.Run("If Delegate.Send fails with an error, SendWithDeadline should return it",
		func(t *testing.T) {
			var (
				wantErr  = errors.New("Delegate.Send error")
				delegate = mock.NewClientDelegate().RegisterSetSendDeadline(
					func(deadline time.Time) (err error) {
						return nil
					},
				).RegisterSend(
					func(seq base.Seq, cmd base.Cmd[any]) (err error) {
						return wantErr
					},
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, err error) {
						err = errors.New("Delegate.Receive error")
						return
					},
				).RegisterClose(
					func() (err error) { return nil },
				)
				mocks  = []*mok.Mock{delegate.Mock}
				client = New[any](delegate, nil)
			)
			_, err := client.SendWithDeadline(time.Time{}, nil, nil)
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
			}
			waitDone(client.Done(), t)
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

	t.Run("Client should forget cmd, if SendWithDeadline failed, because of Delegate.SetSendDeadline",
		func(t *testing.T) {
			var (
				delegate = mock.NewClientDelegate().RegisterSetSendDeadline(
					func(deadline time.Time) (err error) {
						return errors.New("Delegate.SetSendDeadline error")
					},
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, err error) {
						err = errors.New("Delegate.Receive error")
						return
					},
				).RegisterClose(
					func() (err error) { return nil },
				)
				mocks  = []*mok.Mock{delegate.Mock}
				client = New[any](delegate, nil)
			)
			seq, _ := client.SendWithDeadline(time.Time{}, nil, nil)
			if client.Has(seq) {
				t.Error("cmd was not forgotten")
			}
			waitDone(client.Done(), t)
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

	t.Run("Client should forget cmd, if SendWithDeadline failed, because of Delegate.Send",
		func(t *testing.T) {
			var (
				delegate = mock.NewClientDelegate().RegisterSetSendDeadline(
					func(deadline time.Time) (err error) {
						return nil
					},
				).RegisterSend(
					func(seq base.Seq, cmd base.Cmd[any]) (err error) {
						return errors.New("Delegate.Send error")
					},
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, err error) {
						err = errors.New("Delegate.Receive error")
						return
					},
				).RegisterClose(
					func() (err error) { return nil },
				)
				mocks  = []*mok.Mock{delegate.Mock}
				client = New[any](delegate, nil)
			)
			seq, _ := client.SendWithDeadline(time.Time{}, nil, nil)
			if client.Has(seq) {
				t.Error("cmd was not forgotten")
			}
			waitDone(client.Done(), t)
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

	t.Run("We should be able to forget cmd", func(t *testing.T) {
		var (
			cmd         = mock.NewCmd()
			receiveDone = make(chan struct{})
			delegate    = mock.NewClientDelegate().RegisterSend(
				func(seq base.Seq, cmd base.Cmd[any]) (err error) {
					return nil
				},
			).RegisterFlush(
				func() (err error) { return nil },
			).RegisterReceive(
				func() (seq base.Seq, result base.Result, err error) {
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
		seq, _ := client.Send(cmd, nil)
		client.Forget(seq)
		if client.Has(seq) {
			t.Error("cmd was not forgotten")
		}
		close(receiveDone)
		waitDone(client.Done(), t)
		if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
			t.Error(infomap)
		}
	})

	t.Run("If the client was closed or failed to receive a next result, Done channel should be closed and Err method should return the cause",
		func(t *testing.T) {
			var (
				wantErr  = errors.New("Delegate.Receive error")
				delegate = mock.NewClientDelegate().RegisterReceive(
					func() (seq base.Seq, result base.Result, err error) {
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
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
			}
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

	t.Run("We should be able to close the client while it queues a result",
		func(t *testing.T) {
			var (
				wantCmd    = mock.NewCmd()
				wantResult = mock.NewResult().RegisterLastOne(
					func() (lastOne bool) { return true },
				)
				results  = make(chan base.AsyncResult)
				delegate = mock.NewClientDelegate().RegisterSend(
					func(seq base.Seq, cmd base.Cmd[any]) (err error) {
						return errors.New("Delegate.Send error")
					},
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, err error) {
						return 1, wantResult, nil
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
			if err != ErrClosed {
				t.Errorf("unexpected error, want '%v' actual '%v'", ErrClosed, err)
			}
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

	t.Run("If Delegate.Close fails with an error, Close should return it",
		func(t *testing.T) {
			var (
				receiveDone = make(chan struct{})
				wantErr     = errors.New("Delegate.Close error")
				delegate    = mock.NewClientDelegate().RegisterReceive(
					func() (seq base.Seq, result base.Result, err error) {
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
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
			}
			close(receiveDone)
			waitDone(client.Done(), t)
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

	t.Run("If Delegate.Flush fails with an error, Send of all involved Commands should return error",
		func(t *testing.T) {
			var (
				wantErr  = errors.New("flush error")
				cmd1     = mock.NewCmd()
				cmd2     = mock.NewCmd()
				cmd3     = mock.NewCmd()
				delegate = mock.NewClientDelegate().RegisterNSend(3,
					func(seq base.Seq, cmd base.Cmd[any]) (err error) {
						return nil
					},
				).RegisterNFlush(3,
					func() (err error) { return wantErr },
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, err error) {
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
				_, err := client.Send(cmd1, nil)
				if err != wantErr {
					t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
				}
			}()
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := client.Send(cmd2, nil)
				if err != wantErr {
					t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
				}
			}()
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := client.Send(cmd3, nil)
				if err != wantErr {
					t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
				}
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
			// if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
			// 	t.Error(infomap)
			// }
		})

	t.Run("If the client has lost a connection it should try to reconnect",
		func(t *testing.T) {
			var (
				reconected = make(chan struct{})
				delegate   = mock.NewReconnectClientDelegate().RegisterReceive(
					func() (seq base.Seq, result base.Result, err error) {
						err = net.ErrClosed
						return
					},
				).RegisterReconnect(
					func() error { close(reconected); return nil },
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, err error) {
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
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

	t.Run("If the client is closed it should not reconnect", func(t *testing.T) {
		var (
			receiveDone = make(chan struct{})
			delegate    = mock.NewReconnectClientDelegate().RegisterReceive(
				func() (seq base.Seq, result base.Result, err error) {
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
		if err := client.Close(); err != nil {
			t.Errorf("unexpected error, want '%v' actual '%v'", nil, err)
		}
		waitDone(client.Done(), t)
		if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
			t.Error(infomap)
		}
	})

	t.Run("If reconnection fails with an error, it should became the client error",
		func(t *testing.T) {
			var (
				wantErr  = errors.New("reconnection error")
				delegate = mock.NewReconnectClientDelegate().RegisterReceive(
					func() (seq base.Seq, result base.Result, err error) {
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
			if err != wantErr {
				t.Errorf("unexpected error, want '%v' actual '%v'", wantErr, err)
			}
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

	t.Run("Upon creation, the client should call KeepaliveDelegate.Keepalive()",
		func(t *testing.T) {
			var (
				delegate = mock.NewClientKeepaliveDelegate().RegisterKeepalive(
					func(muSn *sync.Mutex) {},
				).RegisterReceive(
					func() (seq base.Seq, result base.Result, err error) {
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
			if infomap := mok.CheckCalls(mocks); len(infomap) > 0 {
				t.Error(infomap)
			}
		})

}

func SameTime(t1, t2 time.Time) bool {
	return !(t1.Before(t2.Truncate(Delta)) || t1.After(t2.Add(Delta)))
}

func waitDone(done <-chan struct{}, t *testing.T) {
	select {
	case <-done:
	case <-time.NewTimer(time.Second).C:
		t.Fatal("test lasts too long")
	}
}
