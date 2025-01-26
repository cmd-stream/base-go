package client

import (
	"context"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cmd-stream/base-go"
)

const (
	inProgress int = iota
	closed
)

// UnexpectedResultCallback is used to process unexpected Results received from
// the server.
//
// This is, when the sequence number of the Result does not match the sequence
// number of any Command sent by the client that is awaiting a Result.
type UnexpectedResultCallback func(seq base.Seq, result base.Result)

// New creates a new Client.
//
// The handler parameter may be nil.
func New[T any](delegate base.ClientDelegate[T],
	callback UnexpectedResultCallback) *Client[T] {
	var (
		ctx, cancel        = context.WithCancel(context.Background())
		flagFl      uint32 = 0
		client             = Client[T]{
			cancel:   cancel,
			delegate: delegate,
			waiting:  make(map[base.Seq]chan<- base.AsyncResult),
			callback: callback,
			done:     make(chan struct{}),
			flagFl:   &flagFl,
			chFl:     make(chan error, 1),
		}
	)
	if keepaliveDelegate, ok := delegate.(base.ClientKeepaliveDelegate[T]); ok {
		keepaliveDelegate.Keepalive(&client.muSn)
	}
	go receive[T](ctx, &client)
	return &client
}

// Client represents a thread-safe, asynchronous cmd-stream client.
//
// It uses ClientDelegate for communication tasks such as sending Commands,
// receiving Results, or managing deadlines. If the connection is lost,
// the client will close, and Client.Err() will provide the connection error.
//
// Client.Close() initiates the process of closing the client. You can track its
// completion by checking Client.Done():
//
//	err = client.Close()
//	if err != nil {
//	  ...
//	}
//	select {
//	case <-time.NewTimer(time.Second).C:
//		err = errors.New("timeout exceeded")
//		...
//	case <-client.Done():
//	}
type Client[T any] struct {
	cancel   context.CancelFunc
	state    int
	delegate base.ClientDelegate[T]
	seq      base.Seq
	waiting  map[base.Seq]chan<- base.AsyncResult
	callback UnexpectedResultCallback
	err      error
	done     chan struct{}
	flagFl   *uint32
	chFl     chan error
	muSn     sync.Mutex
	muWt     sync.Mutex
	muEr     sync.Mutex
	muSt     sync.Mutex
}

// Send sends a Command.
//
// It adds Results received from the server to the results channel. If
// the channel is not large enough, retrieving results for all Commands may hang.
// For each Command, generates a unique sequence number, starting with 1.
// Thus, a Command with seq == 1 is sent first, with seq == 2 is sent second,
// and so on. 0 is reserved for the Ping-Pong game, which keeps a connection
// alive.
//
// Returns the sequence number and an error != nil if the Command was not send.
func (c *Client[T]) Send(cmd base.Cmd[T], results chan<- base.AsyncResult) (
	seq base.Seq, err error) {
	var chFl chan error
	c.muSn.Lock()
	chFl = c.chFl
	c.seq++
	seq = c.seq
	c.memorize(seq, results)
	err = c.delegate.Send(seq, cmd)
	if err != nil {
		c.muSn.Unlock()
		c.Forget(seq)
		return
	}
	c.muSn.Unlock()
	return seq, c.flush(seq, chFl)
}

// SendWithDeadline sends a Command with a deadline.
//
// Use this method if you want to send a Command and specify the deadline.
// In all other it performs like the Send method.
func (c *Client[T]) SendWithDeadline(deadline time.Time, cmd base.Cmd[T],
	results chan<- base.AsyncResult) (seq base.Seq, err error) {
	var chFl chan error
	c.muSn.Lock()
	chFl = c.chFl
	c.seq++
	seq = c.seq
	c.memorize(seq, results)
	err = c.delegate.SetSendDeadline(deadline)
	if err != nil {
		c.muSn.Unlock()
		c.Forget(seq)
		return
	}
	err = c.delegate.Send(seq, cmd)
	if err != nil {
		c.muSn.Unlock()
		c.Forget(seq)
		return
	}
	c.muSn.Unlock()
	return seq, c.flush(seq, chFl)
}

// Has checks if the Command with the specified sequence number has been sent
// by the Client and still waiting for the Result.
func (c *Client[T]) Has(seq base.Seq) bool {
	_, pst := c.load(seq)
	return pst
}

// Forget makes the Client to forget about the Command which still waiting for
// the result.
//
// After calling Forget, all the results of the corresponding Command will be
// handled with UnexpectedResultCallback.
func (c *Client[T]) Forget(seq base.Seq) {
	c.unmemorize(seq)
}

// Done returns a channel that is closed when the Client terminates.
func (c *Client[T]) Done() <-chan struct{} {
	return c.done
}

// Err returns a connection error.
func (c *Client[T]) Err() (err error) {
	c.muEr.Lock()
	err = c.err
	c.muEr.Unlock()
	return
}

// Close terminates the underlying connection and closes the Client.
//
// All Commands waiting for the results will receive an error
// (AsyncResult.Error != nil).
func (c *Client[T]) Close() (err error) {
	c.muSt.Lock()
	defer c.muSt.Unlock()
	c.state = closed
	if err = c.delegate.Close(); err != nil {
		c.state = inProgress
		return
	}
	c.cancel()
	return
}

func (c *Client[T]) receive(ctx context.Context) (err error) {
	defer func() {
		c.unmemorizeAll(err)
	}()
	var (
		seq     base.Seq
		result  base.Result
		results chan<- base.AsyncResult
		pst     bool
	)
	for {
		seq, result, err = c.delegate.Receive()
		if err != nil {
			return
		}
		if result.LastOne() {
			results, pst = c.loadAndUnmemorize(seq)
		} else {
			results, pst = c.load(seq)
		}
		if !pst && c.callback != nil {
			c.callback(seq, result)
			continue
		}
		select {
		case <-ctx.Done():
			return context.Canceled
		case results <- base.AsyncResult{Seq: seq, Result: result}:
			continue
		}
	}
}

func (c *Client[T]) memorize(seq base.Seq, results chan<- base.AsyncResult) {
	c.muWt.Lock()
	c.waiting[seq] = results
	c.muWt.Unlock()
}

func (c *Client[T]) unmemorize(seq base.Seq) {
	c.muWt.Lock()
	delete(c.waiting, seq)
	c.muWt.Unlock()
}

func (c *Client[T]) loadAndUnmemorize(seq base.Seq) (
	results chan<- base.AsyncResult, pst bool) {
	c.muWt.Lock()
	results, pst = c.waiting[seq]
	if pst {
		delete(c.waiting, seq)
	}
	c.muWt.Unlock()
	return
}

func (c *Client[T]) load(seq base.Seq) (results chan<- base.AsyncResult,
	pst bool) {
	c.muWt.Lock()
	results, pst = c.waiting[seq]
	c.muWt.Unlock()
	return
}

func (c *Client[T]) flush(seq base.Seq, chFl chan error) (err error) {
	if swapped := atomic.CompareAndSwapUint32(c.flagFl, 0, 1); !swapped {
		err = <-chFl
		if err != nil {
			chFl <- err
			c.Forget(seq)
		}
		return
	}
	c.muSn.Lock()
	err = c.delegate.Flush()
	if err != nil {
		c.chFl <- err
		c.changeChFl()
		c.muSn.Unlock()
		c.Forget(seq)
		return
	}
	close(c.chFl)
	c.changeChFl()
	c.muSn.Unlock()
	return
}

func (c *Client[T]) changeChFl() {
	c.chFl = make(chan error, 1)
	atomic.CompareAndSwapUint32(c.flagFl, 1, 0)
}

func (c *Client[T]) rangeAndUnmemorize(
	fn func(seq base.Seq, results chan<- base.AsyncResult)) {
	c.muWt.Lock()
	for seq, results := range c.waiting {
		fn(seq, results)
		delete(c.waiting, seq)
	}
	c.muWt.Unlock()
}

func (c *Client[T]) exit(cause error) (err error) {
	c.muSt.Lock()
	if c.state != closed {
		if err = c.delegate.Close(); err != nil {
			c.muSt.Unlock()
			return
		}
	}
	c.muSt.Unlock()

	c.muEr.Lock()
	c.err = cause
	c.muEr.Unlock()
	close(c.done)
	return
}

func (c *Client[T]) unmemorizeAll(cause error) {
	c.rangeAndUnmemorize(func(seq base.Seq, results chan<- base.AsyncResult) {
		queueErrResult(seq, cause, results)
	})
}

func (c *Client[T]) correctErr(err error) error {
	c.muSt.Lock()
	defer c.muSt.Unlock()
	switch c.state {
	case inProgress:
		return err
	case closed:
		return ErrClosed
	default:
		panic("unexpected state")
	}
}

func receive[T any](ctx context.Context, client *Client[T]) {
Start:
	err := client.receive(ctx)
	if err != nil {
		err = client.correctErr(err)
		if netError(err) || err == io.EOF { // TODO Test EOF.
			if reconnectDelegate, ok := client.delegate.(base.ClientReconnectDelegate[T]); ok {
				if err = reconnectDelegate.Reconnect(); err == nil {
					goto Start
				}
			}
		}
	}
	if err = client.exit(err); err != nil {
		panic(err)
	}
}

func queueErrResult(seq base.Seq, err error, results chan<- base.AsyncResult) {
	select {
	case results <- base.AsyncResult{Seq: seq, Error: err}:
	default:
	}
}

func netError(err error) bool {
	_, ok := err.(net.Error)
	return ok
}
