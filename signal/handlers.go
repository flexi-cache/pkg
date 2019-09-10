// Package signal provides an elegant way to handle OS signals which helps modern application to exit gracefully.
package signal

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// DefaultTerminationSignals is used when user doesn't provide their own.
var DefaultTerminationSignals = []os.Signal{syscall.SIGTERM, syscall.SIGINT}

// HandlerFunc is a callback of signal.
type HandlerFunc func(os.Signal)

// Handlers handles OS's signals.
type Handlers struct {
	globalLock            sync.RWMutex
	log                   Logger
	exit                  func(int)
	handlers              map[os.Signal][]HandlerFunc
	terminationSignals    []os.Signal
	terminationProcedures []terminationProcedure
}

type _anySignal struct{}

func (_anySignal) Signal() {}

func (_anySignal) String() string {
	return "AnySignal is not an OS signal, if you see this other than the source code, there is something wrong"
}

var anySignal = _anySignal{}

// NewHandlers creates new Handlers with termination signals set to DefaultTerminationSignals or given signals.
func NewHandlers(terminationSignals ...os.Signal) *Handlers {
	if len(terminationSignals) == 0 {
		terminationSignals = DefaultTerminationSignals
	}
	handlers := &Handlers{
		terminationSignals: terminationSignals,
		handlers:           make(map[os.Signal][]HandlerFunc, len(terminationSignals)+1),
		log:                stdLogger{},
		exit:               os.Exit,
	}
	handlers.handlers[anySignal] = make([]HandlerFunc, 0)
	handlers.installTerminationHandlers()
	return handlers
}

func (s *Handlers) installTerminationHandlers() {
	s.RegisterSignalHandler(s.handleTerminationSignals, s.terminationSignals...)
}

// RegisterSignalHandler registers handler as a callback of all or given signal(s).
// NOTE: if multiple handlers are registered for a single signal, the handlers will be called in registered order, handlers registered to all signals are called first.
func (s *Handlers) RegisterSignalHandler(handler HandlerFunc, signals ...os.Signal) {
	s.globalLock.Lock()
	defer s.globalLock.Unlock()
	if len(signals) == 0 {
		s.handlers[anySignal] = append(s.handlers[anySignal], handler)
		return
	}

	for _, sig := range signals {
		if handlers, exists := s.handlers[sig]; exists {
			s.handlers[sig] = append(handlers, handler)
		} else {
			s.handlers[sig] = []HandlerFunc{handler}
		}
	}
}

// RegisterTerminationProcedure registers given fn as a handler of termination signals, messages are logged before fn called.
func (s *Handlers) RegisterTerminationProcedure(fn TerminationFunc, message string) {
	s.globalLock.Lock()
	s.terminationProcedures = append(s.terminationProcedures, terminationProcedure{fn, message})
	s.globalLock.Unlock()
	s.log.Debug("registered termination procedure for: ", message)
}

// StartListen starts listen to all signals.
// NOTE: termination signals are not required in given signals.
func (s *Handlers) StartListen() context.CancelFunc {
	s.log.Debug("start listening to all signals")
	c := make(chan os.Signal, 1)
	signal.Notify(c)
	go func() {
		for sig := range c {
			s.log.Info("signal received: ", sig)
			s.handleSignal(sig)
		}
	}()
	return func() {
		signal.Stop(c)
		close(c)
	}
}

func (s *Handlers) handleSignal(target os.Signal) {
	s.globalLock.RLock()
	defer s.globalLock.RUnlock()
	for _, handle := range s.handlers[anySignal] {
		handle(target)
	}

	for sig, handlers := range s.handlers {
		if sig != target {
			s.log.Debug("no handler found for signal: ", target)
			continue
		}
		for _, handle := range handlers {
			handle(target)
		}
	}
}

func (s *Handlers) handleTerminationSignals(sig os.Signal) {
	code := s.runTerminationProcedures(sig)
	s.log.Info("bye")
	s.exit(code)
}

func (s *Handlers) runTerminationProcedures(sig os.Signal) int {
	s.globalLock.RLock()
	defer s.globalLock.RUnlock()
	if len(s.terminationProcedures) == 0 {
		s.log.Info("nothing to do before termination")
		return 0
	}

	var code = 0
	for _, proc := range s.terminationProcedures {
		s.log.Info(proc.message)
		err := proc.fn(sig)
		if err == nil {
			continue
		}
		s.log.Info("error while running termination procedure: ", err)
		if code == 0 {
			code = getCodeFromError(err, 1)
		}
	}
	s.log.Info("all termination procedures are done")
	return code
}

func (s *Handlers) setExit(e func(int)) {
	s.exit = e
}

// SetLogger sets the logger to be used.
func (s *Handlers) SetLogger(l Logger) {
	s.globalLock.Lock()
	defer s.globalLock.Unlock()
	s.log = l
}
