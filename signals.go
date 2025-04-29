package gracefulshutdown

import (
	"os"
	"os/signal"
	"slices"
	"syscall"
)

// signalsToListenTo is a list of signals from OS that this package listen to shutdown
var signalsToListenTo = []os.Signal{os.Interrupt, os.Kill, syscall.SIGTERM}

func isShutdownSignal(signal os.Signal) bool {
	return slices.Contains(signalsToListenTo, signal)
}

// newShutdownChannel return a shutdown channel which will be notified on any signal of SignalsToListenTo
func newShutdownChannel() chan os.Signal {
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, signalsToListenTo...)
	return shutdownCh
}
