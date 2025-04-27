package gracefulshutdown_test

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"

	gracefulshutdown "github.com/minhtri06/go-graceful-shutdown"
	"github.com/minhtri06/go-graceful-shutdown/assert"
)

type MockHTTPServer struct {
	listenFunc   func() error
	shutdownFunc func(context.Context) error
}

func (s *MockHTTPServer) ListenAndServe() error {
	return s.listenFunc()
}

func (s *MockHTTPServer) Shutdown(ctx context.Context) error {
	return s.shutdownFunc(ctx)
}

func TestListenAndServe(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		t.Run("if shutdown, should not return error, call ListenAndServe once and Shutdown once", func(t *testing.T) {
			listenCalls := 0
			shutdownCalls := 0
			server := &MockHTTPServer{
				listenFunc: func() error {
					listenCalls++
					return nil
				},
				shutdownFunc: func(context.Context) error {
					shutdownCalls++
					return nil
				},
			}

			shutdown := make(chan os.Signal, 1)
			errChan := make(chan error)
			go func() { errChan <- gracefulshutdown.ListenAndServe(server, shutdown, context.Background()) }()

			time.Sleep(10 * time.Millisecond)
			shutdown <- os.Interrupt
			select {
			case err := <-errChan:
				assert.NoError(t, err)
				assert.Equal(t, listenCalls, 1)
				assert.Equal(t, shutdownCalls, 1)
			case <-time.After(500 * time.Millisecond):
				t.Errorf("timeout waiting for shutdown")
			}
		})

		t.Run("should call Shutdown when shutdown", func(t *testing.T) {
			shutdownCalls := 0
			server := &MockHTTPServer{
				listenFunc: func() error {
					time.Sleep(100 * time.Millisecond)
					return nil
				},
				shutdownFunc: func(context.Context) error {
					shutdownCalls++
					return nil
				},
			}

			shutdown := make(chan os.Signal, 1)
			go gracefulshutdown.ListenAndServe(server, shutdown, context.Background())

			shutdown <- os.Interrupt
			time.Sleep(50 * time.Millisecond)
			assert.Equal(t, shutdownCalls, 1)
		})

		t.Run("if not shutdown, should call ListenAndServe once and not call Shutdown", func(t *testing.T) {
			listenCalls := 0
			shutdownCalls := 0
			server := &MockHTTPServer{
				listenFunc: func() error {
					listenCalls++
					return nil
				},
				shutdownFunc: func(context.Context) error {
					shutdownCalls++
					return nil
				},
			}

			shutdown := make(chan os.Signal)
			go gracefulshutdown.ListenAndServe(server, shutdown, context.Background())

			time.Sleep(50 * time.Millisecond)
			assert.Equal(t, listenCalls, 1)
			assert.Equal(t, shutdownCalls, 0)

			shutdown <- os.Interrupt
			time.Sleep(50 * time.Millisecond)
			assert.Equal(t, listenCalls, 1)
			assert.Equal(t, shutdownCalls, 1)
		})
	})

	t.Run("if ListenAndServe returns error should propagate it", func(t *testing.T) {
		listenErr := errors.New("error when listening")
		server := &MockHTTPServer{
			listenFunc:   func() error { return listenErr },
			shutdownFunc: func(context.Context) error { return nil },
		}

		errChan := make(chan error)
		shutdown := make(chan os.Signal, 1)
		go func() { errChan <- gracefulshutdown.ListenAndServe(server, shutdown, context.Background()) }()

		select {
		case err := <-errChan:
			assert.Error(t, err, listenErr)
		case <-time.After(500 * time.Millisecond):
			t.Errorf("timeout waiting for error to be returned")
		}
	})

	t.Run("should propagate Shutdown's error", func(t *testing.T) {
		shutdownErr := errors.New("error shutting down")
		server := &MockHTTPServer{
			listenFunc: func() error {
				time.Sleep(100 * time.Millisecond)
				return nil
			},
			shutdownFunc: func(context.Context) error { return shutdownErr },
		}

		shutdown := make(chan os.Signal, 1)
		errChan := make(chan error)
		go func() { errChan <- gracefulshutdown.ListenAndServe(server, shutdown, context.Background()) }()
		shutdown <- os.Interrupt

		select {
		case err := <-errChan:
			assert.Error(t, err, shutdownErr)
		case <-time.After(500 * time.Millisecond):
			t.Errorf("timeout waiting for error from Shutdown")
		}
	})

	t.Run("should only shutdown when the signal is os.Interrupt, os.Kill, syscall.SIGINT or syscall.SIGKILL", func(t *testing.T) {
		cases := []struct {
			signal        os.Signal
			shutdownCalls int
		}{
			{syscall.SIGINT, 1},
			{syscall.SIGKILL, 1},
			{os.Interrupt, 1},
			{os.Kill, 1},
			{syscall.SIGABRT, 0},
			{syscall.SIGINFO, 0},
			{syscall.SIGILL, 0},
			{syscall.SIGIOT, 0},
		}

		for _, c := range cases {
			t.Run(c.signal.String(), func(t *testing.T) {
				shutdownCalls := 0
				server := &MockHTTPServer{
					listenFunc: func() error { return nil },
					shutdownFunc: func(ctx context.Context) error {
						shutdownCalls++
						return nil
					},
				}

				shutdown := make(chan os.Signal, 1)
				go gracefulshutdown.ListenAndServe(server, shutdown, context.Background())

				shutdown <- c.signal
				time.Sleep(20 * time.Millisecond)
				assert.Equal(t, shutdownCalls, c.shutdownCalls)

				shutdown <- syscall.SIGINT
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
				shutdownCalls := 0
				server := &MockHTTPServer{
					listenFunc: func() error { return nil },
					shutdownFunc: func(ctx context.Context) error {
						shutdownCalls++
						return nil
					},
				}

				shutdown := make(chan os.Signal, 1)
				go gracefulshutdown.ListenAndServe(server, shutdown, context.Background())

				wrongSignal := syscall.SIGIOT
				for range c.wrongSignalCount {
					shutdown <- wrongSignal
					time.Sleep(20 * time.Millisecond)
					assert.Equal(t, shutdownCalls, 0)
				}

				shutdown <- c.shutdownSignal
				time.Sleep(20 * time.Millisecond)
				assert.Equal(t, shutdownCalls, 1)
			})
		}
	})

	t.Run("should pass the context to the Shutdown function", func(t *testing.T) {
		var gotCtx context.Context
		server := &MockHTTPServer{
			listenFunc: func() error { return nil },
			shutdownFunc: func(ctx context.Context) error {
				gotCtx = ctx
				return nil
			},
		}

		shutdown := make(chan os.Signal, 1)
		type key string
		ctx := context.WithValue(context.Background(), key("test_key"), 12)
		go gracefulshutdown.ListenAndServe(server, shutdown, ctx)

		shutdown <- os.Interrupt
		time.Sleep(10 * time.Millisecond)

		gotVal := gotCtx.Value(key("test_key")).(int)
		assert.Equal(t, gotVal, 12)
	})
}
