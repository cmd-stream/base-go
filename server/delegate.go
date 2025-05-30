package csrv

import (
	"context"
	"net"
)

// Delegate helps the server handle incoming connections.
//
// Handle method should return context.Canceled error if the context is done.
type Delegate interface {
	Handle(ctx context.Context, conn net.Conn) error
}
