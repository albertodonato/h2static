package server

import (
	"crypto/tls"
	"log"
	"net/http"
	"path/filepath"
)

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
