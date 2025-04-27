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

type MockServer struct {
	listenFunc   func() error
	shutdownFunc func(context.Context) error
}

func (s *MockServer) ListenAndServe() error {
	return s.listenFunc()
}

func (s *MockServer) Shutdown(ctx context.Context) error {
	return s.shutdownFunc(ctx)
}

func TestListenAndServe(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		t.Run("if shutdown, should not return error, call ListenAndServe once and Shutdown once", func(t *testing.T) {
			listenCalls := 0
			shutdownCalls := 0
			svr := &MockServer{
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
			go func() {
				time.Sleep(100 * time.Millisecond)
				shutdown <- os.Interrupt
			}()

			err := gracefulshutdown.ListenAndServe(svr, shutdown, context.Background())
			assert.NoError(t, err)
			assert.Equal(t, listenCalls, 1)
			assert.Equal(t, shutdownCalls, 1)
		})

		t.Run("should call Shutdown when shutdown", func(t *testing.T) {
			shutdownCalls := 0
			svr := &MockServer{
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
			go func() {
				gracefulshutdown.ListenAndServe(svr, shutdown, context.Background())
			}()

			shutdown <- os.Interrupt
			time.Sleep(50 * time.Millisecond)
			assert.Equal(t, shutdownCalls, 1)
		})

		t.Run("if not shutdown, should call ListenAndServe once and not call Shutdown", func(t *testing.T) {
			listenCalls := 0
			shutdownCalls := 0
			svr := &MockServer{
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
			go func() {
				gracefulshutdown.ListenAndServe(svr, shutdown, context.Background())
			}()
			time.Sleep(50 * time.Millisecond)
			assert.Equal(t, listenCalls, 1)
			assert.Equal(t, shutdownCalls, 0)
			shutdown <- os.Interrupt
			time.Sleep(50 * time.Millisecond)
			assert.Equal(t, listenCalls, 1)
			assert.Equal(t, shutdownCalls, 1)
		})
	})

	t.Run("should return error if ListenAndServe returns error", func(t *testing.T) {
		listenErr := errors.New("error when listening")
		svr := &MockServer{
			listenFunc:   func() error { return listenErr },
			shutdownFunc: func(context.Context) error { return nil },
		}

		var err error
		shutdown := make(chan os.Signal, 1)
		go func() {
			err = gracefulshutdown.ListenAndServe(svr, shutdown, context.Background())
		}()
		time.Sleep(100 * time.Millisecond)
		assert.Error(t, err, listenErr)
		shutdown <- os.Interrupt
	})

	t.Run("when shutting down, should return an error if Shutdown returns an error", func(t *testing.T) {
		shutdownErr := errors.New("error shutting down")
		svr := &MockServer{
			listenFunc: func() error {
				time.Sleep(100 * time.Millisecond)
				return nil
			},
			shutdownFunc: func(context.Context) error { return shutdownErr },
		}
		shutdown := make(chan os.Signal, 1)
		shutdown <- os.Interrupt
		err := gracefulshutdown.ListenAndServe(svr, shutdown, context.Background())
		assert.Error(t, err, shutdownErr)
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
				server := &MockServer{
					listenFunc: func() error { return nil },
					shutdownFunc: func(ctx context.Context) error {
						shutdownCalls++
						return nil
					},
				}

				shutdown := make(chan os.Signal, 1)
				go func() { gracefulshutdown.ListenAndServe(server, shutdown, context.Background()) }()

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
				server := &MockServer{
					listenFunc: func() error { return nil },
					shutdownFunc: func(ctx context.Context) error {
						shutdownCalls++
						return nil
					},
				}

				shutdown := make(chan os.Signal, 1)
				go func() { gracefulshutdown.ListenAndServe(server, shutdown, context.Background()) }()

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
		server := &MockServer{
			listenFunc: func() error { return nil },
			shutdownFunc: func(ctx context.Context) error {
				gotCtx = ctx
				return nil
			},
		}

		shutdown := make(chan os.Signal, 1)
		ctx := context.WithValue(context.Background(), "test_key", 12)
		go func() { gracefulshutdown.ListenAndServe(server, shutdown, ctx) }()

		shutdown <- os.Interrupt
		time.Sleep(10 * time.Millisecond)

		gotVal := gotCtx.Value("test_key").(int)
		assert.Equal(t, gotVal, 12)
	})
}
