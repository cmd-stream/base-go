package client

import (
	"errors"
)

// ErrClosed happens when the client is closed while connected to the server.
var ErrClosed = errors.New("closed")
