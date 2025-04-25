package gracefulshutdown_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	gracefulshutdown "github.com/minhtri06/go-graceful-shutdown"
	"github.com/minhtri06/go-graceful-shutdown/assertion"
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
			assertion.NoError(t, err)
			assertion.Equal(t, listenCalls, 1)
			assertion.Equal(t, shutdownCalls, 1)
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
			assertion.Equal(t, listenCalls, 1)
			assertion.Equal(t, shutdownCalls, 0)
			shutdown <- os.Interrupt
			time.Sleep(50 * time.Millisecond)
			assertion.Equal(t, listenCalls, 1)
			assertion.Equal(t, shutdownCalls, 1)
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
		assertion.Error(t, err, listenErr)
		shutdown <- os.Interrupt
	})
}
