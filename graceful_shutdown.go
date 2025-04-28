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

// ListenAndServe will call ListenAndServe method from server, and perform graceful shutdown when
// receives a shutdown signal from shutdownCh.
// When shutting down, the Shutdown method of server is called with shutdownCtx as its argument.
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
			if !isShutdownSignal(signal) {
				break
			}
			return server.Shutdown(shutdownCtx)
		}
	}
}
