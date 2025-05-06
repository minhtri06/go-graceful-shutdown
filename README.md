# graceful-shutdown

This is a Go module that would help your HTTP server shutdown gracefully but less confusingly, I guess.

It is part of my practice with TTD (Test-Driven Development) while working through the excellent book [Learn Go with test](https://quii.gitbook.io/learn-go-with-tests) by Chris James. This module is based on an [example](https://github.com/quii/go-graceful-shutdown) from the book's author on github. The idea is simple is simple but testing it wasn't easy for me. However, it has helped me a lot in learning how to write more testable code and get familiar with TDD.

## Install

`go get github.com/minhtri06/graceful-shutdown`

## Usage

Instead of calling `Shutdown` and handle it yourself, you can pass the HTTP server to `gracefulshutdown.ListenAndServe` and let it handle the shutdown for you.

```go
func main() {
    server := &http.Server{
        Addr: ":8000",
        Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            time.Sleep(5 * time.Second)
            w.Write([]byte("You know, I am very slowww..."))
        }),
    }

    shutdownTimeout := 30 * time.Second
    if err := gracefulshutdown.ListenAndServe(server, &shutdownTimeout); err != nil {
        fmt.Printf("error when listening: %v", err)
    }

    fmt.Println("server shutdown gracefully")
}
```
