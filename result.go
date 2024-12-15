package base

// Result represents a command result.
//
// All user-defined results should implement this interface.
type Result interface {
	LastOne() bool
}

// AsyncResult is an asynchronous result, where the Seq is a sequence
// number of the command, Result - a result itself, Error - an error that
// occurred during the command execution.
type AsyncResult struct {
	Seq    Seq
	Result Result
	Error  error
}
