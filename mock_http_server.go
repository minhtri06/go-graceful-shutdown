package gracefulshutdown

import (
	"context"
	"testing"
	"time"
)

type MockHTTPServer struct {
	ListenFunc func() error
	ListenChn  chan struct{}

	ShutdownFunc func(context.Context) error
	ShutdownChn  chan struct{}
}

func NewMockHTTPServer() *MockHTTPServer {
	return &MockHTTPServer{
		ListenFunc:   func() error { return nil },
		ListenChn:    make(chan struct{}, 1),
		ShutdownFunc: func(ctx context.Context) error { return nil },
		ShutdownChn:  make(chan struct{}, 1),
	}
}

func (s *MockHTTPServer) ListenAndServe() error {
	s.ListenChn <- struct{}{}
	return s.ListenFunc()
}

func (s *MockHTTPServer) Shutdown(ctx context.Context) error {
	s.ShutdownChn <- struct{}{}
	return s.ShutdownFunc(ctx)
}

func (s *MockHTTPServer) AssertListenCalled(t testing.TB) {
	t.Helper()
	select {
	case <-s.ListenChn:
	case <-time.After(500 * time.Millisecond):
		t.Errorf("timeout waiting for ListenAndServe to be called")
	}
}

func (s *MockHTTPServer) AssertShutdownCalled(t testing.TB) {
	t.Helper()
	select {
	case <-s.ShutdownChn:
	case <-time.After(500 * time.Millisecond):
		t.Errorf("timeout waiting for Shutdown to be called")
	}
}

func (s *MockHTTPServer) AssertListenNotCalled(t testing.TB) {
	t.Helper()
	select {
	case <-s.ListenChn:
		t.Errorf("expect ListenAndServe not called, but it was called")
	case <-time.After(5 * time.Millisecond):
	}
}

func (s *MockHTTPServer) AssertShutdownNotCalled(t testing.TB) {
	t.Helper()
	select {
	case <-s.ShutdownChn:
		t.Errorf("expect Shutdown not called, but it was called")
	case <-time.After(5 * time.Millisecond):
	}
}
