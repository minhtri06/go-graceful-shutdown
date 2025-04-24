package gracefulshutdown

import (
	"context"
	"os"
)

type HTTPServer interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

func ListenAndServe(server HTTPServer, shutdownSignal chan os.Signal, shutdownCtx context.Context) error {
	return nil
}
