package gracefulshutdown

import (
	"os"
	"os/signal"
	"slices"
	"syscall"
)

// SignalsToListenTo is a list of signals from OS that this package listen to shutdown
var SignalsToListenTo = []os.Signal{os.Interrupt, os.Kill, syscall.SIGTERM}

func isShutdownSignal(signal os.Signal) bool {
	return slices.Contains(SignalsToListenTo, signal)
}

// NewShutdownChannel return a shutdown channel which will be notified on any signal of SignalsToListenTo
func NewShutdownChannel() chan os.Signal {
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, SignalsToListenTo...)
	return shutdownCh
}
