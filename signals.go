package gracefulshutdown

import (
	"os"
	"os/signal"
	"slices"
)

var SignalsToListenTo = []os.Signal{os.Interrupt, os.Kill}

func isShutdownSignal(signal os.Signal) bool {
	return slices.Contains(SignalsToListenTo, signal)
}

func NewShutdownChannel() chan os.Signal {
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, SignalsToListenTo...)
	return shutdownCh
}
