package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/minhtri06/go-graceful-shutdown/cmd"
)

func TestListenAndServe(t *testing.T) {
	const url = "http://localhost:" + port
	cleanup, binPath, err := cmd.BuildBinary(".", "no_graceful")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	interrupt, err := cmd.RunBin(binPath)
	if err != nil {
		t.Fatal(err)
	}

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
		if err != nil {
			t.Error("expect an error but didn't get one")
		}
	case <-time.After(3 * time.Second):
		t.Errorf("timeout waiting for the request error")
	}
}
