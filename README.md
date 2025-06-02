# core-go

[![Go Reference](https://pkg.go.dev/badge/github.com/cmd-stream/core-go.svg)](https://pkg.go.dev/github.com/cmd-stream/core-go)
[![GoReportCard](https://goreportcard.com/badge/cmd-stream/core-go)](https://goreportcard.com/report/github.com/cmd-stream/core-go)
[![codecov](https://codecov.io/gh/cmd-stream/core-go/graph/badge.svg?token=RXPJ6ZIPK7)](https://codecov.io/gh/cmd-stream/core-go)

core-go contains the definitions for both the client and the server.

The client delegates all communication tasks, such as sending commands, 
receiving results, and closing the connection to the `client.Delegate`.

The server uses a configurable number of workers to manage client connections 
via the `server.Delegate`.