package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type cmdFlags struct {
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

const helpHeader = `
Tiny static web server with TLS and HTTP/2 support.

Usage of %s:
`

func parseFlags(fs *flag.FlagSet, args []string) (f cmdFlags, err error) {
	fs.StringVar(&f.Addr, "addr", ":8080", "address and port to listen on")
	fs.StringVar(&f.Dir, "dir", ".", "directory to serve")
	fs.BoolVar(&f.DisableH2, "disable-h2", false, "disable HTTP/2 support")
	fs.BoolVar(&f.Log, "log", false, "log requests")
	fs.StringVar(&f.TLSCert, "tls-cert", "", "certificate file for TLS connections")
	fs.StringVar(&f.TLSKey, "tls-key", "", "key file for TLS connections")
	fs.Usage = func() {
		output := fs.Output()
		fmt.Fprintf(output, helpHeader, fs.Name())
		fs.PrintDefaults()
	}
	err = fs.Parse(args)
	return
}

// Whether to enable TLS.
func enableTLS(flags cmdFlags) bool {
	return flags.TLSCert != "" && flags.TLSKey != ""
}

func runServer(server *http.Server, flags cmdFlags) {
	var err error
	if enableTLS(flags) {
		err = server.ListenAndServeTLS(flags.TLSCert, flags.TLSKey)
	} else {
		err = server.ListenAndServe()
	}
	log.Fatal(err)
}

func setupServer(flags cmdFlags) *http.Server {
	handler := http.FileServer(http.Dir(flags.Dir))
	if flags.Log {
		handler = &loggingHandler{Handler: handler}
	}

	tlsNextProto := map[string]func(*http.Server, *tls.Conn, http.Handler){}
	if !flags.DisableH2 {
		// Setting to nil means to use the default (which is H2-enabled)
		tlsNextProto = nil
	}

	return &http.Server{
		Addr:         flags.Addr,
		Handler:      handler,
		TLSNextProto: tlsNextProto,
	}
}

func main() {
	flags, err := parseFlags(flag.CommandLine, os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
	serveTLS := enableTLS(flags)

	if flags.Log {
		kind := "HTTP"
		if serveTLS {
			kind = "HTTPS"
		}
		absPath, err := filepath.Abs(flags.Dir)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Starting %s server on %s, serving path %s", kind, flags.Addr, absPath)
	}

	server := setupServer(flags)
	runServer(server, flags)
}
