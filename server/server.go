package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/albertodonato/h2static/version"
)

// StaticServerConfig holds configuration options for a StaticServer.
type StaticServerConfig struct {
	Addr                    string
	AllowOutsideSymlinks    bool
	Dir                     string
	DisableH2               bool
	DisableLookupWithSuffix bool
	ShowDotFiles            bool
	Log                     bool
	PasswordFile            string
	TLSCert                 string
	TLSKey                  string
}

// IsHTTPS returns whether HTTPS is enabled in the config.
func (c StaticServerConfig) IsHTTPS() bool {
	return c.TLSCert != "" && c.TLSKey != ""
}

// Validate raises an error if StaticServerConfig is invalid.
func (c StaticServerConfig) Validate() error {
	if err := checkFile(c.Dir, true); err != nil {
		return err
	}

	if c.IsHTTPS() {
		for _, path := range []string{c.TLSCert, c.TLSKey} {
			if err := checkFile(path, false); err != nil {
				return err
			}
		}
	}
	if c.PasswordFile != "" {
		if err := checkFile(c.PasswordFile, false); err != nil {
			return err
		}
	}

	return nil
}

// StaticServer is a static HTTP server.
type StaticServer struct {
	Config StaticServerConfig
}

// NewStaticServer returns a StaticServer.
func NewStaticServer(config StaticServerConfig) (*StaticServer, error) {
	if config.Dir == "" {
		config.Dir = "."
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	server := StaticServer{Config: config}

	// always use absolute path for the root dir.
	absDir, err := filepath.Abs(config.Dir)
	if err != nil {
		return nil, err
	}
	server.Config.Dir = absDir
	return &server, nil
}

// Scheme returns the server scheme (http or https)
func (s *StaticServer) Scheme() string {
	if s.Config.IsHTTPS() {
		return "https"
	}
	return "http"
}

// getServer returns a configured server.
func (s *StaticServer) getServer() (*http.Server, error) {
	fileSystem := FileSystem{
		AllowOutsideSymlinks: s.Config.AllowOutsideSymlinks,
		HideDotFiles:         !s.Config.ShowDotFiles,
		ResolveHTML:          !s.Config.DisableLookupWithSuffix,
		Root:                 s.Config.Dir,
	}
	mux := http.NewServeMux()
	mux.Handle("/", NewFileHandler(fileSystem))
	mux.Handle(
		AssetsPrefix,
		http.StripPrefix(AssetsPrefix, &AssetsHandler{Assets: staticAssets}))

	var handler http.Handler = mux

	if s.Config.PasswordFile != "" {
		credentials, err := loadCredentials(s.Config.PasswordFile)
		if err != nil {
			return nil, err
		}
		handler = &BasicAuthHandler{
			Handler:     handler,
			Credentials: credentials,
			Realm:       version.App.Name,
		}
	}

	if s.Config.Log {
		handler = &LoggingHandler{Handler: handler}
	}
	handler = &CommonHeadersHandler{Handler: handler}

	tlsNextProto := map[string]func(*http.Server, *tls.Conn, http.Handler){}
	if !s.Config.DisableH2 {
		// Setting to nil means to use the default (which is H2-enabled)
		tlsNextProto = nil
	}

	return &http.Server{
		Addr:         s.Config.Addr,
		Handler:      handler,
		TLSNextProto: tlsNextProto,
	}, nil
}

// Run starts the server.
func (s *StaticServer) Run() error {
	if s.Config.Log {
		log.Printf(
			"Starting %s server on %s, serving path %s",
			strings.ToUpper(s.Scheme()), s.Config.Addr, s.Config.Dir)
	}

	return s.runServer()
}

func (s *StaticServer) runServer() error {
	server, err := s.getServer()
	if err != nil {
		return err
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	go func() {
		if s.Config.IsHTTPS() {
			err = server.ListenAndServeTLS(s.Config.TLSCert, s.Config.TLSKey)
		} else {
			err = server.ListenAndServe()
		}
	}()
	<-stop

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if err := server.Shutdown(ctx); err != nil {
		return err
	}
	return s.afterShutdown()
}

func (s *StaticServer) afterShutdown() error {
	log.Printf("Server shutdown")
	return nil
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
