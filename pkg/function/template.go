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
	"strings"
	"time"

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

	fn := func(w http.ResponseWriter, r *http.Request) {
		// If we get requests from the kubelet's prober logic, handle it so
		// the user function doesn't have to.
		if strings.HasPrefix(r.Header.Get("User-Agent"), "kube-probe/") {
			select {
				// If we've received the termination signal, then start
				// to fail readiness probes to drain traffic away from
				// this replica.
				case <-ctx.Done():
					http.Error(w, "shutting down", http.StatusServiceUnavailable)

				// Otherwise, start to pass readiness probes as soon as
				// we are able to invoke the user function.
				default:
					w.WriteHeader(http.StatusOK)
			}
		} else {
			// If there is no kubelet probe header, then pass requests along
			// to the user function.
			p.{{.Function}}(w, r)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{
		Addr: fmt.Sprint(":", port),
		Handler: http.HandlerFunc(fn),
	}

	// Start the server in a go routine, so that we can monitor for shutdown and gracefully
	// drain requests on the main thread.
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// When the context is cancelled we got a SIGTERM, which means that we will start to
	// fail readiness probes, and as this is observed by the orchestration layer, traffic
	// will migrate elsewhere.
	<-ctx.Done()

	// After a grace period, stop accepting new connections, and wait for outstanding
	// requests to complete.
	time.Sleep(30 * time.Second)

	// Don't use ctx because it is already closed!
	log.Fatal(server.Shutdown(context.Background()))
}
`

var mainTemplate = template.Must(template.New("http-go-function-main").Parse(packageMain))
