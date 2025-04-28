package gracefulshutdown

import (
	"context"
	"os"
)

type HTTPServer interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

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
