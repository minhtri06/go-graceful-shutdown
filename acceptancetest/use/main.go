package main

import (
	"fmt"
	"net/http"
	"time"

	gracefulshutdown "github.com/minhtri06/graceful-shutdown"
	"github.com/minhtri06/graceful-shutdown/acceptancetest"
)

const port = "8000"

func main() {
	server := &http.Server{
		Addr:    ":" + port,
		Handler: http.HandlerFunc(acceptancetest.SlowHandler),
	}

	shutdownTimeout := 30 * time.Second
	if err := gracefulshutdown.ListenAndServe(server, &shutdownTimeout); err != nil {
		fmt.Printf("error when listening: %v", err)
	}

	fmt.Println("server shutdown gracefully")
}
