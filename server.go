package gracefulshutdown

import (
	"context"
	"os"
)

type HTTPServer interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

func ListenAndServe(server HTTPServer, shutdownChan chan os.Signal, shutdownCtx context.Context) error {
	listenErr := make(chan error)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			listenErr <- err
		}
	}()
	select {
	case err := <-listenErr:
		return err
	case <-shutdownChan:
		if err := server.Shutdown(shutdownCtx); err != nil {
			return err
		}
	}
	return nil
}
