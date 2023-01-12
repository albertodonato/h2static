package server

import (
	_ "expvar"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
)

// serveDebug serves /debug URLs on the specified port on localhost.
func serveDebug(port uint) {
	err := http.ListenAndServe(fmt.Sprintf("localhost:%d", port), http.DefaultServeMux)
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
