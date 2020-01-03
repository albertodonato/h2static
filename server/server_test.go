package server_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/albertodonato/h2static/server"
)

func TestServer(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}

type ServerTestSuite struct {
	TempDirTestSuite
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

// getServer returns a configured http.Server
func (s *ServerTestSuite) TestGetServerDefault() {
	serv := server.StaticServer{}
	httpServer := server.GetServer(serv)
	s.Equal("", httpServer.Addr)
	s.IsType(http.FileServer(http.Dir(".")), httpServer.Handler)
	s.Nil(httpServer.TLSNextProto)
}

// getServer returns a configured http.Server with the specified dir
func (s *ServerTestSuite) TestSetupServerSpecifyDir() {
	serv := server.StaticServer{Dir: "/some/dir"}
	httpServer := server.GetServer(serv)
	s.Equal("", httpServer.Addr)
	s.IsType(http.FileServer(http.Dir("/some/dir")), httpServer.Handler)
}

// getServer returns a configured http.Server with logging
func (s *ServerTestSuite) TestSetupServerLog() {
	serv := server.StaticServer{Log: true}
	httpServer := server.GetServer(serv)
	s.IsType(&server.LoggingHandler{}, httpServer.Handler)
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
	s.IsType(&server.BasicAuthHandler{}, httpServer.Handler)
}
