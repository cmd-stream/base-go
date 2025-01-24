# base-go

[![Go Reference](https://pkg.go.dev/badge/github.com/cmd-stream/base-go.svg)](https://pkg.go.dev/github.com/cmd-stream/base-go)
[![GoReportCard](https://goreportcard.com/badge/cmd-stream/base-go)](https://goreportcard.com/report/github.com/cmd-stream/base-go)
[![codecov](https://codecov.io/gh/cmd-stream/base-go/graph/badge.svg?token=RXPJ6ZIPK7)](https://codecov.io/gh/cmd-stream/base-go)

base-go contains the definitions for both the client and server.

The client delegates all communication tasks, such as sending Commands,
receiving Results, and closing the connection, to the `ClientDelegate`.

The server utilizes a configurable number of workers to manage client 
connections using `ServerDelegate`.