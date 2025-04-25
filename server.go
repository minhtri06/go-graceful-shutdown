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
	if err := server.ListenAndServe(); err != nil {
		return err
	}
	<-shutdownChan
	server.Shutdown(shutdownCtx)
	return nil
}
