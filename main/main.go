package main

import (
	"context"
	"fmt"
	"net/http"

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

	if err := gracefulshutdown.ListenAndServe(httpServer, context.Background()); err != nil {
		fmt.Println("error")
	}
}
