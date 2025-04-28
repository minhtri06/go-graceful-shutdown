package gracefulshutdown_test

import (
	"context"
	"testing"
	"time"
)

type MockHTTPServer struct {
	ListenFunc func() error
	ListenCh   chan struct{}

	ShutdownFunc func(context.Context) error
	ShutdownCh   chan struct{}
}

func NewMockHTTPServer() *MockHTTPServer {
	return &MockHTTPServer{
		ListenFunc:   func() error { return nil },
		ListenCh:     make(chan struct{}, 1),
		ShutdownFunc: func(ctx context.Context) error { return nil },
		ShutdownCh:   make(chan struct{}, 1),
	}
}

func (s *MockHTTPServer) ListenAndServe() error {
	s.ListenCh <- struct{}{}
	return s.ListenFunc()
}

func (s *MockHTTPServer) Shutdown(ctx context.Context) error {
	s.ShutdownCh <- struct{}{}
	return s.ShutdownFunc(ctx)
}

func (s *MockHTTPServer) AssertListenCalled(t testing.TB) {
	t.Helper()
	select {
	case <-s.ListenCh:
	case <-time.After(500 * time.Millisecond):
		t.Errorf("timeout waiting for ListenAndServe to be called")
	}
}

func (s *MockHTTPServer) AssertShutdownCalled(t testing.TB) {
	t.Helper()
	select {
	case <-s.ShutdownCh:
	case <-time.After(500 * time.Millisecond):
		t.Errorf("timeout waiting for Shutdown to be called")
	}
}

func (s *MockHTTPServer) AssertListenNotCalled(t testing.TB) {
	t.Helper()
	select {
	case <-s.ListenCh:
		t.Errorf("expect ListenAndServe not called, but it was called")
	case <-time.After(5 * time.Millisecond):
	}
}

func (s *MockHTTPServer) AssertShutdownNotCalled(t testing.TB) {
	t.Helper()
	select {
	case <-s.ShutdownCh:
		t.Errorf("expect Shutdown not called, but it was called")
	case <-time.After(5 * time.Millisecond):
	}
}
