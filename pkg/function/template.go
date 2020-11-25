package function

import "text/template"

const packageMain = `
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

        p "{{.Package}}"
)

func main() {
	// When we get a SIGTERM, cancel the context.
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-c
		cancel()
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{
		Addr: fmt.Sprint(":", port),
		Handler: http.HandlerFunc(p.{{.Function}}),
	}

	// Start the server in a go routine, so that we can monitor for shutdown and gracefully
	// drain requests on the main thread.
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// When the context is cancelled we got a SIGTERM, so drain HTTP requests.
	// Don't use ctx because it is already closed!
	// TODO(mattmoor): Add a built-in readiness probe handler that starts failing when
	// ctx is cancelled for ~30s before calling Shutdown.
	<-ctx.Done()
	log.Fatal(server.Shutdown(context.Background()))
}
`

var mainTemplate = template.Must(template.New("http-go-function-main").Parse(packageMain))
