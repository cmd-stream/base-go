package base

// Result represents a result of the cmd-stream command.
//
// All user-defined results should implement this interface. If the LastOne
// method returns false, the client will wait more command results.
type Result interface {
	LastOne() bool
}

// AsyncResult is an asynchronous result, where the Seq field is a sequence
// number of the command, Result field - a result itself, Error field - an
// error that occurred during command execution.
type AsyncResult struct {
	Seq    Seq
	Result Result
	Error  error
}
