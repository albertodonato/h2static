package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
)

type flags struct {
	Addr      string
	Dir       string
	DisableH2 bool
	Log       bool
	TLSCert   string
	TLSKey    string
}

type loggingHandler struct {
	Handler http.Handler
}

func (lh *loggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s", r.Proto, r.Method, r.URL)
	lh.Handler.ServeHTTP(w, r)
}

func parseFlags() (f flags) {
	flag.StringVar(&f.Addr, "addr", ":8080", "address and port to listen on")
	flag.StringVar(&f.Dir, "dir", ".", "directory to serve")
	flag.BoolVar(&f.DisableH2, "disable-h2", false, "disable HTTP/2 support")
	flag.BoolVar(&f.Log, "log", false, "log requests")
	flag.StringVar(&f.TLSCert, "tls-cert", "", "certificate file for TLS connections")
	flag.StringVar(&f.TLSKey, "tls-key", "", "key file for TLS connections")
	flag.CommandLine.Usage = func() {
		f := flag.CommandLine
		output := f.Output()
		fmt.Fprintf(output, "Tiny static web server with TLS and HTTP/2 support.\n\n")
		fmt.Fprintf(output, "Usage of %s:\n", f.Name())
		f.PrintDefaults()
	}
	flag.Parse()
	return
}

func main() {
	flags := parseFlags()
	serveTLS := flags.TLSCert != "" && flags.TLSKey != ""

	if flags.Log {
		kind := "HTTP"
		if serveTLS {
			kind = "HTTPS"
		}
		log.Printf("Starting %s server on %s, serving path %s", kind, flags.Addr, flags.Dir)
	}

	handler := http.FileServer(http.Dir(flags.Dir))
	if flags.Log {
		handler = &loggingHandler{Handler: handler}
	}

	tlsNextProto := map[string]func(*http.Server, *tls.Conn, http.Handler){}
	if !flags.DisableH2 {
		// Setting to nil means to use the default (which is H2-enabled)
		tlsNextProto = nil
	}

	server := http.Server{
		Addr:         flags.Addr,
		Handler:      handler,
		TLSNextProto: tlsNextProto,
	}
	var err error
	if serveTLS {
		err = server.ListenAndServeTLS(flags.TLSCert, flags.TLSKey)
	} else {
		err = server.ListenAndServe()
	}
	log.Fatal(err)
}
