package main

import (
	"fmt"
	"net/http"

	"github.com/minhtri06/graceful-shutdown/acceptancetest"
)

const port = "8000"

func main() {
	server := &http.Server{
		Addr:    ":" + port,
		Handler: http.HandlerFunc(acceptancetest.SlowHandler),
	}

	if err := server.ListenAndServe(); err != nil {
		fmt.Println(err)
	}
}
