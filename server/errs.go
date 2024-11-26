package server

import "errors"

// ErrNoWorkers happens when the Server configured with 0 workers.
var ErrNoWorkers = errors.New("not positive Conf.WorkersCount")

// ErrNotServing happens when the Server is not serving and is closed or
// shutdown.
var ErrNotServing = errors.New("server is not serving")

// ErrShutdown happens when the Server is shutdown while serving.
var ErrShutdown = errors.New("shutdown")

// ErrClosed happens when the Server is closed while serving.
var ErrClosed = errors.New("closed")
