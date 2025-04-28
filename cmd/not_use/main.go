package main

import (
	"fmt"
	"net/http"

	"github.com/minhtri06/go-graceful-shutdown/cmd"
)

const port = "8000"

func main() {
	server := &http.Server{
		Addr:    ":" + port,
		Handler: http.HandlerFunc(cmd.SlowHandler),
	}

	if err := server.ListenAndServe(); err != nil {
		fmt.Println(err)
	}
}
