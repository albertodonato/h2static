package server

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/albertodonato/h2static/version"
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

// ValidateConfig validates the StaticServer configuration
func (s StaticServer) ValidateConfig() error {
	if err := checkFile(s.Dir, true); err != nil {
		return err
	}

	if s.IsHTTPS() {
		for _, path := range []string{s.TLSCert, s.TLSKey} {
			if err := checkFile(path, false); err != nil {
				return err
			}
		}
	}
	if s.PasswordFile != "" {
		if err := checkFile(s.PasswordFile, false); err != nil {
			return err
		}
	}

	return nil
}

// IsHTTPS returns whether HTTPS is enabled.
func (s StaticServer) IsHTTPS() bool {
	return s.TLSCert != "" && s.TLSKey != ""
}

// getServer returns a configured server.
func (s StaticServer) getServer() (*http.Server, error) {
	fileSystem := NewFileSystem(
		s.Dir, !s.DisableLookupWithSuffix, !s.ShowDotFiles)
	mux := http.NewServeMux()
	mux.Handle("/", NewFileHandler(fileSystem))
	mux.Handle(
		AssetsPrefix,
		http.StripPrefix(AssetsPrefix, &AssetsHandler{Assets: staticAssets}))

	var handler http.Handler = mux

	if s.PasswordFile != "" {
		credentials, err := loadCredentials(s.PasswordFile)
		if err != nil {
			return nil, err
		}
		handler = &BasicAuthHandler{
			Handler:     handler,
			Credentials: credentials,
			Realm:       version.App.Name,
		}
	}

	if s.Log {
		handler = &LoggingHandler{Handler: handler}
	}
	handler = &CommonHeadersHandler{Handler: handler}

	tlsNextProto := map[string]func(*http.Server, *tls.Conn, http.Handler){}
	if !s.DisableH2 {
		// Setting to nil means to use the default (which is H2-enabled)
		tlsNextProto = nil
	}

	return &http.Server{
		Addr:         s.Addr,
		Handler:      handler,
		TLSNextProto: tlsNextProto,
	}, nil
}

// Run starts the server.
func (s StaticServer) Run() error {
	var err error

	server, err := s.getServer()
	if err != nil {
		return err
	}
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

func checkFile(path string, asDir bool) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	isDir := info.IsDir()
	if asDir && !isDir {
		return fmt.Errorf("not a directory: %s", path)
	}
	if !asDir && isDir {
		return fmt.Errorf("is a directory: %s", path)
	}
	return nil
}
