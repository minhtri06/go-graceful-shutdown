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

// ListenAndServe will call server.ListenAndServe method, and perform graceful shutdown when
// an interrupt or kill signal sent to shutdownCh.
// shutdownTimeout is the timeout for shutting down, if it's nil then there's no timeout.
func ListenAndServe(server HTTPServer, shutdownCh chan os.Signal, shutdownTimeout *time.Duration) error {
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
}
