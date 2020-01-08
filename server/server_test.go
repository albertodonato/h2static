package server_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/albertodonato/h2static/server"
	"github.com/albertodonato/h2static/testhelpers"
)

func TestServer(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}

type ServerTestSuite struct {
	testhelpers.TempDirTestSuite
}

// If the specified Dir doesn't exist, an error is returned.
func (s *ServerTestSuite) TestValidateConfigDirNotExists() {
	serv := server.StaticServer{Dir: "/not/here"}
	err := serv.ValidateConfig()
	s.NotNil(err)
	s.Contains(err.Error(), "/not/here: no such file or directory")
}

// If the specified Dir exists but is not a directory, an error is returned.
func (s *ServerTestSuite) TestValidateConfigDirNotDir() {
	path := s.WriteFile("foo", "bar")
	serv := server.StaticServer{Dir: path}
	err := serv.ValidateConfig()
	s.NotNil(err)
	s.Equal(fmt.Sprintf("not a directory: %s", path), err.Error())
}

// If the TLS certificate file doesn't exist, an error is returned.
func (s *ServerTestSuite) TestValidateConfigTLSCertFileNotExists() {
	tlsKey := s.WriteFile("foo", "bar")
	serv := server.StaticServer{Dir: s.TempDir, TLSCert: "/not/here", TLSKey: tlsKey}
	err := serv.ValidateConfig()
	s.NotNil(err)
	s.Contains(err.Error(), "/not/here: no such file or directory")
}

// If the TLS key file doesn't exist, an error is returned.
func (s *ServerTestSuite) TestValidateConfigTLSKeyFileNotExists() {
	tlsCert := s.WriteFile("foo", "bar")
	serv := server.StaticServer{Dir: s.TempDir, TLSCert: tlsCert, TLSKey: "/not/here"}
	err := serv.ValidateConfig()
	s.NotNil(err)
	s.Contains(err.Error(), "/not/here: no such file or directory")
}

// If the passwords file doesn't exist, an error is returned.
func (s *ServerTestSuite) TestValidateConfigPasswordFileNotExists() {
	serv := server.StaticServer{Dir: s.TempDir, PasswordFile: "/not/here"}
	err := serv.ValidateConfig()
	s.NotNil(err)
	s.Contains(err.Error(), "/not/here: no such file or directory")
}

// If no invalid file is passed, ValidateConfig returns nil.
func (s *ServerTestSuite) TestValidateConfigNoError() {
	serv := server.StaticServer{Dir: s.TempDir}
	s.Nil(serv.ValidateConfig())
}

// IsHTTPS returns true if certificates are set.
func (s *ServerTestSuite) TestEnableTLSTrue() {
	serv := server.StaticServer{
		TLSCert: "cert",
		TLSKey:  "secret",
	}
	s.True(serv.IsHTTPS())
}

// IsHTTPS returns false if certificates are not set.
func (s *ServerTestSuite) TestEnableTLSFalse() {
	serv := server.StaticServer{}
	s.False(serv.IsHTTPS())
}

// getServer returns static file handlers for a path.
func (s *ServerTestSuite) TestGetServerDefaultStaticServe() {
	serv := server.StaticServer{}
	httpServer, err := server.GetServer(serv)
	s.Nil(err)
	s.Nil(httpServer.TLSNextProto)
	s.Equal("", httpServer.Addr)
	h := httpServer.Handler.(*server.CommonHeadersHandler)
	mux := h.Handler.(*http.ServeMux)
	fh, pattern := mux.Handler(httptest.NewRequest("GET", "/foo", nil))
	s.Equal("/", pattern)
	s.IsType(&server.FileHandler{}, fh)
	s.IsType(".", fh.(*server.FileHandler).FileSystem.Root)
}

// getServer returns a configured http.Server with the specified dir
func (s *ServerTestSuite) TestSetupServerSpecifyDir() {
	serv := server.StaticServer{Dir: "/some/dir"}
	httpServer, err := server.GetServer(serv)
	s.Nil(err)
	h := httpServer.Handler.(*server.CommonHeadersHandler)
	mux := h.Handler.(*http.ServeMux)
	fh, _ := mux.Handler(httptest.NewRequest("GET", "/foo", nil))
	s.IsType("/some/dir", fh.(*server.FileHandler).FileSystem.Root)
}

// getServer returns a configured http.Server with logging
func (s *ServerTestSuite) TestSetupServerLog() {
	serv := server.StaticServer{Log: true}
	httpServer, err := server.GetServer(serv)
	s.Nil(err)
	h := httpServer.Handler.(*server.CommonHeadersHandler)
	s.IsType(&server.LoggingHandler{}, h.Handler)
}

// getServer returns a configured http.Server without HTTP/2
func (s *ServerTestSuite) TestSetupServerNoH2() {
	serv := server.StaticServer{DisableH2: true}
	httpServer, err := server.GetServer(serv)
	s.Nil(err)
	s.NotNil(httpServer.TLSNextProto)
}

// getServer returns a configured http.Server with Basic-Auth
func (s *ServerTestSuite) TestSetupServerBasicAuth() {
	absPath := s.WriteFile("basic-auth", "")
	serv := server.StaticServer{PasswordFile: absPath}
	httpServer, err := server.GetServer(serv)
	s.Nil(err)
	h := httpServer.Handler.(*server.CommonHeadersHandler)
	s.IsType(&server.BasicAuthHandler{}, h.Handler)
}
