package gracefulshutdown

import (
	"os"
	"os/signal"
	"slices"
)

var ShutdownSignalsListenTo = []os.Signal{os.Interrupt, os.Kill}

func isShutdownSignal(signal os.Signal) bool {
	return slices.Contains(ShutdownSignalsListenTo, signal)
}

func NewShutdownChannel() chan os.Signal {
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, ShutdownSignalsListenTo...)
	return shutdownCh
}
