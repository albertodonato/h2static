// Package server provides an HTTP server for serving static files.
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
	"strconv"
	"strings"
	"time"

	"github.com/albertodonato/h2static/version"
)

// StaticServerConfig holds configuration options for a StaticServer.
type StaticServerConfig struct {
	Addr                    string
	AllowOutsideSymlinks    bool
	CSS                     string
	DebugAddr               string
	Dir                     string
	DisableH2               bool
	DisableIndex            bool
	DisableLookupWithSuffix bool
	Log                     bool
	PasswordFile            string
	RequestPathPrefix       string
	ShowDotFiles            bool
	TLSCert                 string
	TLSKey                  string
}

// Port returns the port from the config.
func (c StaticServerConfig) Port() uint16 {
	i := strings.LastIndex(c.Addr, ":")
	n, err := strconv.ParseUint(c.Addr[i+1:], 10, 16)
	if err != nil {
		return 0
	}
	return uint16(n)
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
	if c.CSS != "" {
		if err := checkFile(c.CSS, false); err != nil {
			return err
		}
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
	mux := http.NewServeMux()
	// handler for static files
	fileSystem := FileSystem{
		AllowOutsideSymlinks: s.Config.AllowOutsideSymlinks,
		HideDotFiles:         !s.Config.ShowDotFiles,
		ResolveHTML:          !s.Config.DisableLookupWithSuffix,
		Root:                 s.Config.Dir,
	}
	mux.Handle("/", NewFileHandler(fileSystem, !s.Config.DisableIndex, s.Config.RequestPathPrefix))

	// add handler for builtin assets. Cache them for 24h so they don't
	// get requested every time
	assetsHandler := AddHeadersHandler(
		map[string]string{"Cache-Control": fmt.Sprintf("public, max-age=%d", 24*60*60)},
		AssetsHandler(),
	)
	mux.Handle(AssetsPrefix, http.StripPrefix(AssetsPrefix, assetsHandler))

	// optionally, serve CSS from the specified file instead of the builtin assets
	if s.Config.CSS != "" {
		mux.HandleFunc(
			CSSAsset,
			func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, s.Config.CSS)
			})
	}

	var handler http.Handler = mux
	// optionally, strip request path prefix
	if s.Config.RequestPathPrefix != "" {
		handler = http.StripPrefix(s.Config.RequestPathPrefix, handler)
	}
	// optionally, wrap handler with Basic-Auth
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

	// optionally, enable logging
	if s.Config.Log {
		handler = &LoggingHandler{Handler: handler}
	}

	// always add server version to headers
	handler = AddHeadersHandler(
		map[string]string{"Server": version.App.Identifier()},
		handler,
	)

	tlsNextProto := make(map[string]func(*http.Server, *tls.Conn, http.Handler))
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
	log.Printf("Starting %v %s server on %s, serving path %s",
		version.App, strings.ToUpper(s.Scheme()), s.Config.Addr, s.Config.Dir)

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
		var err error
		if s.Config.IsHTTPS() {
			err = server.ListenAndServeTLS(s.Config.TLSCert, s.Config.TLSKey)
		} else {
			err = server.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	if s.Config.DebugAddr != "" {
		log.Printf("Serving debug URLs on %s", s.Config.DebugAddr)

		var handler http.Handler = newDebugMux()
		// optionally, enable logging
		if s.Config.Log {
			handler = &LoggingHandler{Handler: handler}
		}
		// always add server version to headers
		handler = AddHeadersHandler(
			map[string]string{"Server": version.App.Identifier()},
			handler,
		)

		go func() {
			debugServer := http.Server{
				Addr:    s.Config.DebugAddr,
				Handler: handler,
			}
			if err := debugServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}()
	}
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
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
