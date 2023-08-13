package server

import (
	"net"
	"time"
)

// LostConnCallback is called when Server loses connection with the client.
type LostConnCallback = func(addr net.Addr, err error)

// Conf is a base Server configuration.
//
// Workerscount parameter must > 0. LostConnCallback may be nil.
type Conf struct {
	ConnReceiverConf
	WorkersCount     int
	LostConnCallback LostConnCallback
}

// ConnReceiverConf is a ConnReceiver configuration.
//
// FirstConnTimeout defines the time during which Server should accept the
// first connection, if == 0 waits forever.
type ConnReceiverConf struct {
	FirstConnTimeout time.Duration
}
