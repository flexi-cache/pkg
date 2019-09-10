package signal_test

import (
	"os"
	"syscall"

	"github.com/flexi-cache/pkg/signal"
)

func ExampleHandlers() {
	signals := signal.NewHandlers()
	stop := signals.StartListen()
	defer stop()

	var cleanup = func(os.Signal) error {
		// This will be called when termination signals received. e.g. SIGINT
		return nil
	}

	signals.RegisterTerminationProcedure(cleanup, "cleaning something up")

	var doSomething = func(os.Signal) {
		// This will be called when SIGUSR1 or SIGUSR2 received.
	}
	signals.RegisterSignalHandler(doSomething, syscall.SIGUSR1, syscall.SIGUSR2)
	//Output:
}
