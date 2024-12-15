# base-go

[![Go Reference](https://pkg.go.dev/badge/github.com/cmd-stream/base-go.svg)](https://pkg.go.dev/github.com/cmd-stream/base-go)
[![GoReportCard](https://goreportcard.com/badge/cmd-stream/base-go)](https://goreportcard.com/report/github.com/cmd-stream/base-go)
[![codecov](https://codecov.io/gh/cmd-stream/base-go/graph/badge.svg?token=RXPJ6ZIPK7)](https://codecov.io/gh/cmd-stream/base-go)

base-go serves as the foundation for building the cmd-stream-go client and 
server.

The client delegates communication with the server to `ClientDelegate`, which 
handles sending commands and receiving results. Similarly, the server uses 
`ServerDelegate` to manage connections, execute commands, and send back results.

# Tests
Test coverage is about 93%.