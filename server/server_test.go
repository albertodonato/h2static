package server_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/albertodonato/h2static/server"
	"github.com/albertodonato/h2static/testhelpers"
)

func TestStaticServerConfig(t *testing.T) {
	suite.Run(t, new(StaticServerConfigTestSuite))
}

type StaticServerConfigTestSuite struct {
	testhelpers.TempDirTestSuite
}

// If the specified Dir doesn't exist, an error is returned.
func (s *StaticServerConfigTestSuite) TestConfigValidateDirNotExists() {
	config := server.StaticServerConfig{Dir: "/not/here"}
	err := config.Validate()
	s.NotNil(err)
	s.Contains(err.Error(), "/not/here: no such file or directory")
}

// If the specified Dir exists but is not a directory, an error is returned.
func (s *StaticServerConfigTestSuite) TestConfigValidateDirNotDir() {
	path := s.WriteFile("foo", "bar")
	config := server.StaticServerConfig{Dir: path}
	err := config.Validate()
	s.NotNil(err)
	s.Equal(fmt.Sprintf("not a directory: %s", path), err.Error())
}

// If the TLS certificate file doesn't exist, an error is returned.
func (s *StaticServerConfigTestSuite) TestConfigValidateTLSCertFileNotExists() {
	tlsKey := s.WriteFile("foo", "bar")
	config := server.StaticServerConfig{
		Dir:     s.TempDir,
		TLSCert: "/not/here",
		TLSKey:  tlsKey,
	}
	err := config.Validate()
	s.NotNil(err)
	s.Contains(err.Error(), "/not/here: no such file or directory")
}

// If the TLS key file doesn't exist, an error is returned.
func (s *StaticServerConfigTestSuite) TestConfigValidateTLSKeyFileNotExists() {
	tlsCert := s.WriteFile("foo", "bar")
	config := server.StaticServerConfig{
		Dir:     s.TempDir,
		TLSCert: tlsCert,
		TLSKey:  "/not/here",
	}
	err := config.Validate()
	s.NotNil(err)
	s.Contains(err.Error(), "/not/here: no such file or directory")
}

// If the passwords file doesn't exist, an error is returned.
func (s *StaticServerConfigTestSuite) TestConfigValidatePasswordFileNotExists() {
	config := server.StaticServerConfig{
		Dir:          s.TempDir,
		PasswordFile: "/not/here",
	}
	err := config.Validate()
	s.NotNil(err)
	s.Contains(err.Error(), "/not/here: no such file or directory")
}

// If no invalid file is passed, ValidateConfig returns nil.
func (s *StaticServerConfigTestSuite) TestConfigValidateNoError() {
	config := server.StaticServerConfig{Dir: s.TempDir}
	s.Nil(config.Validate())
}

// IsHTTPS returns false if certificates are not set in the config.
func (s *StaticServerConfigTestSuite) TestIsHTTPSFalse() {
	config := server.StaticServerConfig{
		Dir:     s.TempDir,
		TLSCert: "cert.pem",
		TLSKey:  "key.pem",
	}
	s.True(config.IsHTTPS())
}

// IsHTTPS returns true if certificates are set in the config.
func (s *StaticServerConfigTestSuite) TestIsHTTPSTrue() {
	config := server.StaticServerConfig{Dir: s.TempDir}
	s.False(config.IsHTTPS())
}

// Port returns the service port from the config.
func (s *StaticServerConfigTestSuite) TestPort() {
	config := server.StaticServerConfig{Addr: "localhost:1234"}
	s.Equal(config.Port(), uint16(1234))
}

// Port returns 0 if the address is unset.
func (s *StaticServerConfigTestSuite) TestPortUnset() {
	config := server.StaticServerConfig{}
	s.Equal(config.Port(), uint16(0))
}

func TestStaticServer(t *testing.T) {
	suite.Run(t, new(StaticServerTestSuite))
}

type StaticServerTestSuite struct {
	testhelpers.TempDirTestSuite
}

// If certificates are set and exist, HTTPS is enabled.
func (s *StaticServerTestSuite) TestEnableTLSTrue() {
	tlsCert := s.WriteFile("cert.pem", "cert")
	tlsKey := s.WriteFile("key.pem", "key")
	serv, err := server.NewStaticServer(server.StaticServerConfig{
		Dir:     s.TempDir,
		TLSCert: tlsCert,
		TLSKey:  tlsKey,
	})
	s.Nil(err)
	s.True(serv.Config.IsHTTPS())
	s.Equal(serv.Scheme(), "https")
}

// If certificates are not set, HTTPS is not enabled.
func (s *StaticServerTestSuite) TestEnableTLSFalse() {
	serv := server.StaticServer{}
	s.False(serv.Config.IsHTTPS())
	s.Equal(serv.Scheme(), "http")
}

// getServer returns static file handlers for a path.
func (s *StaticServerTestSuite) TestGetServerDefaultStaticServe() {
	serv, err := server.NewStaticServer(server.StaticServerConfig{})
	s.Nil(err)
	httpServer, err := server.GetServer(serv)
	s.Nil(err)
	s.Nil(httpServer.TLSNextProto)
	s.Equal("", httpServer.Addr)
	h := httpServer.Handler.(*server.CommonHeadersHandler)
	mux := h.Handler.(*http.ServeMux)
	fh, pattern := mux.Handler(httptest.NewRequest("GET", "/foo", nil))
	s.Equal("/", pattern)
	s.IsType(&server.FileHandler{}, fh)
	curdir, err := os.Getwd()
	s.Nil(err)
	s.Equal(curdir, fh.(*server.FileHandler).FileSystem.Root)
}

// getServer returns a configured http.Server with the specified dir
func (s *StaticServerTestSuite) TestSetupServerSpecifyDir() {
	serv, err := server.NewStaticServer(server.StaticServerConfig{Dir: s.TempDir})
	s.Nil(err)
	httpServer, err := server.GetServer(serv)
	s.Nil(err)
	h := httpServer.Handler.(*server.CommonHeadersHandler)
	mux := h.Handler.(*http.ServeMux)
	fh, _ := mux.Handler(httptest.NewRequest("GET", "/foo", nil))
	s.Equal(s.TempDir, fh.(*server.FileHandler).FileSystem.Root)
}

// getServer returns a configured http.Server which allows outside symlink access
func (s *StaticServerTestSuite) TestSetupServerSpecifyAllowOutsideSymlinks() {
	serv, err := server.NewStaticServer(server.StaticServerConfig{
		Dir:                  s.TempDir,
		AllowOutsideSymlinks: true,
	})
	s.Nil(err)
	httpServer, err := server.GetServer(serv)
	s.Nil(err)
	h := httpServer.Handler.(*server.CommonHeadersHandler)
	mux := h.Handler.(*http.ServeMux)
	fh, _ := mux.Handler(httptest.NewRequest("GET", "/foo", nil))
	s.True(fh.(*server.FileHandler).FileSystem.AllowOutsideSymlinks)
}

// getServer returns a configured http.Server with logging
func (s *StaticServerTestSuite) TestSetupServerLog() {
	serv, err := server.NewStaticServer(server.StaticServerConfig{
		Dir: s.TempDir,
		Log: true,
	})
	s.Nil(err)
	httpServer, err := server.GetServer(serv)
	s.Nil(err)
	h := httpServer.Handler.(*server.CommonHeadersHandler)
	s.IsType(&server.LoggingHandler{}, h.Handler)
}

// getServer returns a configured http.Server without HTTP/2
func (s *StaticServerTestSuite) TestSetupServerNoH2() {
	serv, err := server.NewStaticServer(server.StaticServerConfig{
		Dir:       s.TempDir,
		DisableH2: true,
	})
	s.Nil(err)
	httpServer, err := server.GetServer(serv)
	s.Nil(err)
	s.NotNil(httpServer.TLSNextProto)
}

// getServer returns a configured http.Server with Basic-Auth
func (s *StaticServerTestSuite) TestSetupServerBasicAuth() {
	absPath := s.WriteFile("basic-auth", "")
	serv, err := server.NewStaticServer(server.StaticServerConfig{
		Dir:          s.TempDir,
		PasswordFile: absPath,
	})
	s.Nil(err)
	httpServer, err := server.GetServer(serv)
	s.Nil(err)
	h := httpServer.Handler.(*server.CommonHeadersHandler)
	s.IsType(&server.BasicAuthHandler{}, h.Handler)
}
