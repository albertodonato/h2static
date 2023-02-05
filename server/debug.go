package server

import (
	"expvar"
	"fmt"
	"net/http"
	"net/http/pprof"
)

// newDebugServer returns a configured debug Server on localhost.
func newDebugServer(port uint) *http.Server {
	return &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", port),
		Handler: newDebugMux(),
	}
}

// newDebugMux returns a new ServeMux configured with debug URLs.
func newDebugMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/vars", expvar.Handler())
	return mux
}
