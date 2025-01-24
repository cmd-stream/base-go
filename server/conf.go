package server

import (
	"time"
)

// Conf is a base Server configuration.
//
// WorkersCount parameter must > 0. LostConnCallback may be nil.
type Conf struct {
	ConnReceiverConf
	WorkersCount int
}

// ConnReceiverConf configures a ConnReceiver.
//
// FirstConnTimeout specifies the time within which the Server must accept the
// first connection. If set to 0, it will wait indefinitely.
type ConnReceiverConf struct {
	FirstConnTimeout time.Duration
}
