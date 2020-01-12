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

// StaticServer is a static HTTP server.
type StaticServer struct {
	Config StaticServerConfig
}

// NewStaticServer returns a StaticServer.
func NewStaticServer(config StaticServerConfig) (*StaticServer, error) {
	if config.Dir == "" {
		config.Dir = "."
	}
	server := StaticServer{Config: config}
	if err := validateServerConfig(server); err != nil {
		return nil, err
	}

	// always use absolute path for the root dir.
	absDir, err := filepath.Abs(config.Dir)
	if err != nil {
		return nil, err
	}
	server.Config.Dir = absDir
	return &server, nil
}

func validateServerConfig(s StaticServer) error {
	if err := checkFile(s.Config.Dir, true); err != nil {
		return err
	}

	if s.IsHTTPS() {
		for _, path := range []string{s.Config.TLSCert, s.Config.TLSKey} {
			if err := checkFile(path, false); err != nil {
				return err
			}
		}
	}
	if s.Config.PasswordFile != "" {
		if err := checkFile(s.Config.PasswordFile, false); err != nil {
			return err
		}
	}

	return nil
}

// IsHTTPS returns whether HTTPS is enabled.
func (s *StaticServer) IsHTTPS() bool {
	return s.Config.TLSCert != "" && s.Config.TLSKey != ""
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
	var err error

	server, err := s.getServer()
	if err != nil {
		return err
	}
	isHTTPS := s.IsHTTPS()
	if s.Config.Log {
		kind := "HTTP"
		if isHTTPS {
			kind = "HTTPS"
		}
		log.Printf(
			"Starting %s server on %s, serving path %s",
			kind, s.Config.Addr, s.Config.Dir)
	}

	if isHTTPS {
		err = server.ListenAndServeTLS(s.Config.TLSCert, s.Config.TLSKey)
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
