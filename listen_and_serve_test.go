package gracefulshutdown

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/minhtri06/graceful-shutdown/assert"
)

func TestListenAndServe(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		t.Run("if shutdown, should not return error, call ListenAndServe once and Shutdown once", func(t *testing.T) {
			server := NewMockHTTPServer()

			shutdown := make(chan os.Signal, 1)
			errCh := make(chan error)
			go func() { errCh <- listenAndServe(server, shutdown, nil) }()

			shutdown <- os.Interrupt
			select {
			case err := <-errCh:
				assert.NoError(t, err)
				server.AssertListenCalled(t)
				server.AssertShutdownCalled(t)
			case <-time.After(500 * time.Millisecond):
				t.Errorf("timeout waiting for shutdown")
			}
		})

		t.Run("if not shutdown, should call ListenAndServe once and do not call Shutdown", func(t *testing.T) {
			server := NewMockHTTPServer()

			shutdown := make(chan os.Signal)
			go listenAndServe(server, shutdown, nil)

			time.Sleep(100 * time.Millisecond) // wait sometime
			server.AssertListenCalled(t)
			server.AssertShutdownNotCalled(t)

			shutdown <- os.Interrupt
			server.AssertListenNotCalled(t) // not called again
			server.AssertShutdownCalled(t)
		})
	})

	t.Run("if ListenAndServe returns error should propagate it", func(t *testing.T) {
		listenErr := errors.New("error when listening")
		server := NewMockHTTPServer()
		server.ListenFunc = func() error { return listenErr }

		errCh := make(chan error)
		shutdown := make(chan os.Signal, 1)
		go func() { errCh <- listenAndServe(server, shutdown, nil) }()

		select {
		case err := <-errCh:
			assert.Error(t, err, listenErr)
			server.AssertListenCalled(t)
			server.AssertShutdownNotCalled(t)
		case <-time.After(500 * time.Millisecond):
			t.Errorf("timeout waiting for error to be returned")
		}
	})

	t.Run("should propagate Shutdown's error", func(t *testing.T) {
		shutdownErr := errors.New("error shutting down")
		server := NewMockHTTPServer()
		server.ShutdownFunc = func(ctx context.Context) error { return shutdownErr }

		shutdown := make(chan os.Signal, 1)
		errCh := make(chan error)
		go func() { errCh <- listenAndServe(server, shutdown, nil) }()

		shutdown <- os.Interrupt
		select {
		case err := <-errCh:
			assert.Error(t, err, shutdownErr)
			server.AssertListenCalled(t)
			server.AssertShutdownCalled(t)
		case <-time.After(500 * time.Millisecond):
			t.Errorf("timeout waiting for error from Shutdown")
		}
	})

	t.Run("should only shutdown when receive related signals", func(t *testing.T) {
		cases := []struct {
			signal         os.Signal
			shutdownCalled bool
		}{
			{syscall.SIGINT, true},
			{syscall.SIGKILL, true},
			{syscall.SIGTERM, true},
			{os.Interrupt, true},
			{os.Kill, true},
			{syscall.SIGABRT, false},
			{syscall.SIGINFO, false},
			{syscall.SIGILL, false},
			{syscall.SIGIOT, false},
		}

		for _, c := range cases {
			t.Run(c.signal.String(), func(t *testing.T) {
				server := NewMockHTTPServer()

				shutdown := make(chan os.Signal, 1)
				go listenAndServe(server, shutdown, nil)

				shutdown <- c.signal
				if c.shutdownCalled {
					server.AssertShutdownCalled(t)
				} else {
					server.AssertShutdownNotCalled(t)
				}
			})
		}
	})

	t.Run("after send a wrong signal, if we send a actual shutdown signal, it should shutdown", func(t *testing.T) {
		cases := []struct {
			wrongSignalCount int
			shutdownSignal   os.Signal
		}{
			{4, os.Interrupt},
			{6, os.Kill},
			{7, syscall.SIGINT},
			{1, syscall.SIGKILL},
		}
		for _, c := range cases {
			t.Run(c.shutdownSignal.String(), func(t *testing.T) {
				server := NewMockHTTPServer()

				shutdown := make(chan os.Signal, 1)
				go listenAndServe(server, shutdown, nil)

				wrongSignal := syscall.SIGIOT
				for range c.wrongSignalCount {
					shutdown <- wrongSignal
					server.AssertShutdownNotCalled(t)
				}

				shutdown <- c.shutdownSignal
				server.AssertShutdownCalled(t)
			})
		}
	})

	t.Run("should timeout if the Shutdown method takes too long to complete", func(t *testing.T) {
		server := NewMockHTTPServer()
		timeout := 20 * time.Millisecond
		server.ShutdownFunc = func(ctx context.Context) error {
			select {
			case <-ctx.Done():
			case <-time.After(22 * time.Millisecond):
				t.Errorf("expect to timeout")
			}
			return nil
		}

		shutdown := make(chan os.Signal, 1)
		endCh := make(chan error, 1)
		go func() { endCh <- listenAndServe(server, shutdown, &timeout) }()

		shutdown <- os.Interrupt

		select {
		case <-endCh:
		case <-time.After(500 * time.Millisecond):
			t.Errorf("timeout waiting for context")
		}
	})

	t.Run("should not timeout if Shutdown complete earlier timeout", func(t *testing.T) {
		server := NewMockHTTPServer()
		timeout := 20 * time.Millisecond
		server.ShutdownFunc = func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				t.Errorf("didn't expect to timeout")
			case <-time.After(15 * time.Millisecond):
			}
			return nil
		}

		shutdown := make(chan os.Signal, 1)
		endCh := make(chan error, 1)
		go func() { endCh <- listenAndServe(server, shutdown, &timeout) }()

		shutdown <- os.Interrupt

		select {
		case <-endCh:
		case <-time.After(500 * time.Millisecond):
			t.Errorf("timeout waiting for context")
		}
	})
}
