package gracefulshutdown

import "context"

type HTTPServer interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

func ListenAndServe(server HTTPServer, shutdownCtx context.Context) error {
	return nil
}
