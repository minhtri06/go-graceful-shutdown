package gracefulshutdown_test

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"

	graceshut "github.com/minhtri06/go-graceful-shutdown"
	"github.com/minhtri06/go-graceful-shutdown/assert"
)

func TestListenAndServe(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		t.Run("if shutdown, should not return error, call ListenAndServe once and Shutdown once", func(t *testing.T) {
			server := NewMockHTTPServer()

			shutdown := make(chan os.Signal, 1)
			errCh := make(chan error)
			go func() { errCh <- graceshut.ListenAndServe(server, shutdown, nil) }()

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

		t.Run("if not shutdown, should call ListenAndServe once and not call Shutdown", func(t *testing.T) {
			server := NewMockHTTPServer()

			shutdown := make(chan os.Signal)
			go graceshut.ListenAndServe(server, shutdown, nil)

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
		go func() { errCh <- graceshut.ListenAndServe(server, shutdown, nil) }()

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
		go func() { errCh <- graceshut.ListenAndServe(server, shutdown, nil) }()

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
				go graceshut.ListenAndServe(server, shutdown, nil)

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
				go graceshut.ListenAndServe(server, shutdown, nil)

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

	t.Run("obey the timeout in the context", func(t *testing.T) {
		server := NewMockHTTPServer()
		timeout := 40 * time.Millisecond
		server.ShutdownFunc = func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				// If it goes here, the timeout is less than 20 ms
				t.Errorf("expect to ctx timeout more than 20 ms")
			case <-time.After(20 * time.Millisecond):
			}
			return nil
		}

		shutdown := make(chan os.Signal, 1)
		endCh := make(chan error, 1)
		go func() { endCh <- graceshut.ListenAndServe(server, shutdown, &timeout) }()

		time.Sleep(40 * time.Millisecond)
		shutdown <- os.Interrupt

		select {
		case <-endCh:
		case <-time.After(3000 * time.Millisecond):
			t.Errorf("timeout waiting for context")
		}
	})
}
