package signal

import "os"

// TerminationFunc is a callback of termination signals.
// NOTE: if error is not nil, it will be logged out, and the exit code of the whole process will be non zero.
// Use WrapErrorWithCode to specify the desired exit code or 1 will be returned.
// The first error returning from TerminationFunc determines the exit code.
type TerminationFunc func(os.Signal) error

// NewTerminationFunc creates a TerminationFunc without parameter nor return value.
func NewTerminationFunc(fn func()) TerminationFunc {
	return func(os.Signal) error {
		fn()
		return nil
	}
}

type terminationProcedure struct {
	fn      TerminationFunc
	message string
}

type errorWithExitCode struct {
	error
	exitCode int
}

func getCodeFromError(e error, def int) int {
	if ee, ok := e.(errorWithExitCode); ok {
		return ee.exitCode
	}
	return def
}

// WrapErrorWithCode wraps given error with exit code, which is useful while returning from a TerminationFunc.
func WrapErrorWithCode(e error, code int) error {
	if e == nil {
		return nil
	}
	return errorWithExitCode{e, code}
}
