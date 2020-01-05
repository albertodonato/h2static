package server_test

import (
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
	httpServer := server.GetServer(serv)
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
	httpServer := server.GetServer(serv)
	h := httpServer.Handler.(*server.CommonHeadersHandler)
	mux := h.Handler.(*http.ServeMux)
	fh, _ := mux.Handler(httptest.NewRequest("GET", "/foo", nil))
	s.IsType("/some/dir", fh.(*server.FileHandler).FileSystem.Root)
}

// getServer returns a configured http.Server with logging
func (s *ServerTestSuite) TestSetupServerLog() {
	serv := server.StaticServer{Log: true}
	httpServer := server.GetServer(serv)
	h := httpServer.Handler.(*server.CommonHeadersHandler)
	s.IsType(&server.LoggingHandler{}, h.Handler)
}

// getServer returns a configured http.Server without HTTP/2
func (s *ServerTestSuite) TestSetupServerNoH2() {
	serv := server.StaticServer{DisableH2: true}
	httpServer := server.GetServer(serv)
	s.NotNil(httpServer.TLSNextProto)
}

// getServer returns a configured http.Server with Basic-Auth
func (s *ServerTestSuite) TestSetupServerBasicAuth() {
	absPath := s.WriteFile("basic-auth", "")
	serv := server.StaticServer{PasswordFile: absPath}
	httpServer := server.GetServer(serv)
	h := httpServer.Handler.(*server.CommonHeadersHandler)
	s.IsType(&server.BasicAuthHandler{}, h.Handler)
}
