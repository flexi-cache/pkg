package signal

import (
	"io"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func _newHandlers(exit func(int)) *Handlers {
	handlers := NewHandlers()
	if exit == nil {
		exit = func(int) {}
	}
	handlers.setExit(exit)
	return handlers
}

func TestHandlersExecutionOrderOK(t *testing.T) {
	t.Parallel()
	handlers := NewHandlers(nil)

	called := ""
	fnAll1 := func(os.Signal) {
		called += "a1 "
	}
	fnAll2 := func(os.Signal) {
		called += "a2 "
	}
	fn1 := func(os.Signal) {
		called += "1 "
	}
	fn2 := func(os.Signal) {
		called += "2 "
	}

	handlers.RegisterSignalHandler(fn1, syscall.SIGUSR1)
	handlers.RegisterSignalHandler(fn2, syscall.SIGUSR1)
	handlers.RegisterSignalHandler(fnAll1)
	handlers.RegisterSignalHandler(fnAll2)
	handlers.handleSignal(syscall.SIGUSR1)
	assert.Equal(t, "a1 a2 1 2 ", called)
}

func TestHandlersTerminationProceduresExecutionOrderOK(t *testing.T) {
	t.Parallel()
	handlers := NewHandlers()
	order := ""
	fn1 := func(os.Signal) error {
		order += "1 "
		return nil
	}
	fn2 := func(os.Signal) error {
		order += "2 "
		return nil
	}

	handlers.RegisterTerminationProcedure(fn1, "")
	handlers.RegisterTerminationProcedure(fn2, "")
	handlers.runTerminationProcedures(syscall.SIGINT)
	assert.Equal(t, "1 2 ", order)
}

func TestHandlersEmptyTerminationProceduresNotPanics(t *testing.T) {
	t.Parallel()
	handlers := NewHandlers()
	assert.NotPanics(t, func() {
		handlers.runTerminationProcedures(syscall.SIGINT)
	})
}

func TestHandlersHandlesDefaultTerminationSignals(t *testing.T) {
	t.Parallel()
	exitCalled := 0
	handlers := _newHandlers(func(int) { exitCalled++ })

	for i, sig := range DefaultTerminationSignals {
		handlers.handleSignal(sig)
		assert.Equal(t, i+1, exitCalled)
	}
}

func TestHandlersListensForExpectedSignals(t *testing.T) {
	handlers := _newHandlers(nil)

	var called = ""

	handlers.RegisterSignalHandler(func(sig os.Signal) {
		called += sig.String()
	}, syscall.SIGUSR1)

	handlers.RegisterTerminationProcedure(func(sig os.Signal) error {
		called += sig.String()
		return nil
	}, "")

	stop := handlers.StartListen()
	defer stop()

	for _, sig := range append(DefaultTerminationSignals, syscall.SIGUSR1) {
		syscall.Kill(os.Getpid(), sig.(syscall.Signal))
		time.Sleep(time.Millisecond * 50)
		assert.Contains(t, called, sig.String())
	}
}

func TestHandlersReturnsZeroWhileNoError(t *testing.T) {
	t.Parallel()
	var ret = -1
	handlers := _newHandlers(func(code int) {
		ret = code
	})

	handlers.handleSignal(syscall.SIGTERM)
	assert.Equal(t, 0, ret)

	handlers.RegisterTerminationProcedure(func(os.Signal) error {
		return nil
	}, "")
	handlers.handleSignal(syscall.SIGTERM)
	assert.Equal(t, 0, ret)
}

func TestHandlersReturnsOneWhileError(t *testing.T) {
	t.Parallel()
	var ret = -1
	handlers := _newHandlers(func(code int) {
		ret = code
	})

	handlers.RegisterTerminationProcedure(func(os.Signal) error {
		return io.EOF
	}, "")
	handlers.handleSignal(syscall.SIGTERM)
	assert.Equal(t, 1, ret)
}

func TestHandlersReturnsGivenCodeWhileErrorSpecified(t *testing.T) {
	t.Parallel()
	var ret = -1
	handlers := _newHandlers(func(code int) {
		ret = code
	})

	handlers.RegisterTerminationProcedure(func(os.Signal) error {
		return WrapErrorWithCode(io.EOF, 42)
	}, "")
	handlers.handleSignal(syscall.SIGTERM)
	assert.Equal(t, 42, ret)
}

func TestHandlersSetLoggerOK(t *testing.T) {
	t.Parallel()
	handlers := _newHandlers(nil)
	oldLogger := handlers.log
	handlers.SetLogger(nil)
	newLogger := handlers.log
	assert.Equal(t, nil, newLogger)
	assert.NotEqual(t, newLogger, oldLogger)
}

func TestSignalStopDoesNotClosesTheChan(t *testing.T) {
	t.Parallel()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	signal.Stop(c)
	assert.NotPanics(t, func() {
		close(c)
	})
}

func TestAnySignalContainsWarning(t *testing.T) {
	assert.Contains(t, anySignal.String(),
		"if you see this other than the source code, there is something wrong")
}
