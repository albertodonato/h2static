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

const helpHeader = `
Tiny static web server with TLS and HTTP/2 support.

Usage of %s:
`

type LoggingHandler struct {
	Handler http.Handler
}

func (lh LoggingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s %s %s", r.Proto, r.Method, r.URL)
	lh.Handler.ServeHTTP(w, r)
}

type StaticServer struct {
	Addr      string
	Dir       string
	DisableH2 bool
	Log       bool
	TLSCert   string
	TLSKey    string
}

// NewStaticServerFromCmdline returns a a StaticServer parsing cmdline args.
func NewStaticServerFromCmdline(fs *flag.FlagSet, args []string) (*StaticServer, error) {
	s := &StaticServer{}
	fs.StringVar(&s.Addr, "addr", ":8080", "address and port to listen on")
	fs.StringVar(&s.Dir, "dir", ".", "directory to serve")
	fs.BoolVar(&s.DisableH2, "disable-h2", false, "disable HTTP/2 support")
	fs.BoolVar(&s.Log, "log", false, "log requests")
	fs.StringVar(&s.TLSCert, "tls-cert", "", "certificate file for TLS connections")
	fs.StringVar(&s.TLSKey, "tls-key", "", "key file for TLS connections")
	fs.Usage = func() {
		output := fs.Output()
		fmt.Fprintf(output, helpHeader, fs.Name())
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return s, nil
}

// isHTTPS returns whether HTTPS is enabled.
func (s StaticServer) isHTTPS() bool {
	return s.TLSCert != "" && s.TLSKey != ""
}

// getServer returns a configured server.
func (s StaticServer) getServer() *http.Server {
	handler := http.FileServer(http.Dir(s.Dir))
	if s.Log {
		handler = &LoggingHandler{Handler: handler}
	}

	tlsNextProto := map[string]func(*http.Server, *tls.Conn, http.Handler){}
	if !s.DisableH2 {
		// Setting to nil means to use the default (which is H2-enabled)
		tlsNextProto = nil
	}

	return &http.Server{
		Addr:         s.Addr,
		Handler:      handler,
		TLSNextProto: tlsNextProto,
	}
}

// Run starts the server.
func (s StaticServer) Run() error {
	var err error

	server := s.getServer()
	isHTTPS := s.isHTTPS()
	if s.Log {
		kind := "HTTP"
		if isHTTPS {
			kind = "HTTPS"
		}
		absPath, err := filepath.Abs(s.Dir)
		if err != nil {
			return err
		}
		log.Printf("Starting %s server on %s, serving path %s", kind, s.Addr, absPath)
	}

	if isHTTPS {
		err = server.ListenAndServeTLS(s.TLSCert, s.TLSKey)
	} else {
		err = server.ListenAndServe()
	}
	return err
}

func main() {
	server, err := NewStaticServerFromCmdline(flag.CommandLine, os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
