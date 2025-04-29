package gracefulshutdown

import (
	"context"
	"os"
	"time"
)

// HTTPServer is an abstraction for something that listen for connection and do HTTP works.
type HTTPServer interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

// ListenAndServe will call server.ListenAndServe method and perform graceful shutdown when notified
// a interrupt or kill signal from the OS.
// When shutting down, server.Shutdown method will be called with the context timeout of shutdownTimeout
func ListenAndServe(server HTTPServer, shutdownTimeout *time.Duration) error {
	return listenAndServe(server, newShutdownChannel(), shutdownTimeout)
}

func listenAndServe(server HTTPServer, shutdownCh chan os.Signal, shutdownTimeout *time.Duration) error {
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
				continue
			}
			ctx := context.Background()
			if shutdownTimeout != nil {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), *shutdownTimeout)
				defer cancel()
			}
			return server.Shutdown(ctx)
		}
	}
}
