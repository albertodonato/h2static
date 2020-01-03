package server

import (
	"crypto/tls"
	"log"
	"net/http"
	"path/filepath"
)

// StaticServer is a static HTTP server.
type StaticServer struct {
	Addr                    string
	Dir                     string
	DisableH2               bool
	DisableLookupWithSuffix bool
	ShowDotFiles            bool
	Log                     bool
	PasswordFile            string
	TLSCert                 string
	TLSKey                  string
}

// IsHTTPS returns whether HTTPS is enabled.
func (s StaticServer) IsHTTPS() bool {
	return s.TLSCert != "" && s.TLSKey != ""
}

// getServer returns a configured server.
func (s StaticServer) getServer() *http.Server {
	fileSystem := FileSystem{
		FileSystem:   http.Dir(s.Dir),
		ResolveHTML:  !s.DisableLookupWithSuffix,
		HideDotFiles: !s.ShowDotFiles,
	}
	handler := http.FileServer(fileSystem)
	if s.PasswordFile != "" {
		credentials, err := loadCredentials(s.PasswordFile)
		if err != nil {
			panic(err)
		}
		handler = &BasicAuthHandler{
			Handler:     handler,
			Credentials: credentials,
			Realm:       "h2static",
		}
	}

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
	isHTTPS := s.IsHTTPS()
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
