package csrv

import (
	"net"
	"reflect"
	"testing"
	"time"
)

func TestOptions(t *testing.T) {
	var (
		o                                     = Options{}
		wantWorkersCount                      = 1
		wantLostConnCallback LostConnCallback = func(addr net.Addr, err error) {}
		wantConnReceiver                      = []SetConnReceiverOption{}
	)
	Apply([]SetOption{
		WithWorkersCount(wantWorkersCount),
		WithLostConnCallback(wantLostConnCallback),
		WithConnReceiver(wantConnReceiver...),
	}, &o)

	if o.WorkersCount != wantWorkersCount {
		t.Errorf("unexpected WorkersCount, want %v actual %v", wantWorkersCount,
			o.WorkersCount)
	}

	if o.LostConnCallback == nil {
		t.Errorf("LostConnCallback == nil")
	}

	if !reflect.DeepEqual(o.ConnReceiver, wantConnReceiver) {
		t.Errorf("unexpected ConnReceiver, want %v actual %v", wantConnReceiver,
			o.ConnReceiver)
	}

}

func TestConnReceiverOptions(t *testing.T) {
	var (
		o                    = ConnReceiverOptions{}
		wantFirstConnTimeout = time.Second
	)
	ApplyForConnReceiver([]SetConnReceiverOption{
		WithFirstConnTimeout(wantFirstConnTimeout),
	}, &o)

	if o.FirstConnTimeout != wantFirstConnTimeout {
		t.Errorf("unexpected FirstConnTimeout, want %v actual %v",
			wantFirstConnTimeout, o.FirstConnTimeout)
	}

}
