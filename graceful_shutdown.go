package gracefulshutdown

import (
	"context"
	"os"
)

// HTTPServer is an abstraction for something that listen for connection and do HTTP works.
type HTTPServer interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

// ListenAndServe will call server.ListenAndServe method, and perform graceful shutdown when
// a signal that included in [SignalsToListenTo] sent to shutdownCh channel.
// When shutting down, server.Shutdown method is called with shutdownCtx as its argument.
func ListenAndServe(server HTTPServer, shutdownCh chan os.Signal, shutdownCtx context.Context) error {
	listenErr := make(chan error)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			listenErr <- err
		}
	}()
	for {
		select {
		case err := <-listenErr:
			return err
		case signal := <-shutdownCh:
			if isShutdownSignal(signal) {
				return server.Shutdown(shutdownCtx)
			}
		}
	}
}
