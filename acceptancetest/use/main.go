package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"

	gracefulshutdown "github.com/minhtri06/go-graceful-shutdown"
	"github.com/minhtri06/go-graceful-shutdown/acceptancetest"
)

const port = "8000"

func main() {
	server := &http.Server{
		Addr:    ":" + port,
		Handler: http.HandlerFunc(acceptancetest.SlowHandler),
	}

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt)

	if err := gracefulshutdown.ListenAndServe(server, nil); err != nil {
		fmt.Printf("error when listening: %v", err)
	}

	fmt.Println("server shutdown gracefully")
}
