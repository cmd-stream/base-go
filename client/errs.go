package bcln

import (
	"errors"
)

// ErrClosed happens when the Client is closed while connected to the server.
var ErrClosed = errors.New("closed")
