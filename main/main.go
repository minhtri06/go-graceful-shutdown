package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	gracefulshutdown "github.com/minhtri06/go-graceful-shutdown"
)

func routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	}))
	return mux
}

func main() {
	httpServer := &http.Server{
		Addr:    ":8000",
		Handler: routes(),
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
	if err := gracefulshutdown.ListenAndServe(httpServer, shutdown, context.Background()); err != nil {
		fmt.Println("error")
	}
}
