package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/minhtri06/go-graceful-shutdown/acceptancetest"
	"github.com/minhtri06/go-graceful-shutdown/assert"
)

func TestListenAndServe(t *testing.T) {
	const url = "http://localhost:" + port

	cleanup, interrupt, err := acceptancetest.RunServer(".", port)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	_, err = http.Get(url)
	if err != nil {
		t.Errorf("expect no error, but got %v", err)
	}

	// Fire up a request, we know it's slow, so without graceful shutdown it would return error
	errCh := make(chan error, 1)
	go func() {
		_, err := http.Get(url)
		errCh <- err
	}()

	// Wait a moment for the request actually hits the server,
	// otherwise we stop the server before the request to hit
	time.Sleep(50 * time.Millisecond)
	if err := interrupt(); err != nil {
		t.Fatal(err)
	}

	select {
	case err := <-errCh:
		assert.AnyError(t, err)
	case <-time.After(3 * time.Second):
		t.Errorf("timeout waiting for the request error")
	}
}
