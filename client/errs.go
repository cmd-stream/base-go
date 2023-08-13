package client

import (
	"errors"
)

// ErrClosed happens when Client is closed while connected to the server.
var ErrClosed = errors.New("closed")
