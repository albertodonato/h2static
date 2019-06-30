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
	suite.Suite
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
	s.Equal(httpServer.Addr, "")
	s.IsType(httpServer.Handler, http.FileServer(http.Dir(".")))
	s.Nil(httpServer.TLSNextProto)
}

// getServer returns a configured http.Server with the specified dir
func (s *ServerTestSuite) TestSetupServerSpecifyDir() {
	serv := server.StaticServer{Dir: "/some/dir"}
	httpServer := server.GetServer(serv)
	s.Equal(httpServer.Addr, "")
	s.IsType(httpServer.Handler, http.FileServer(http.Dir("/some/dir")))
}

// getServer returns a configured http.Server with logging
func (s *ServerTestSuite) TestSetupServerLog() {
	serv := server.StaticServer{Log: true}
	httpServer := server.GetServer(serv)
	s.IsType(httpServer.Handler, &server.LoggingHandler{})
}

// getServer returns a configured http.Server without HTTP/2
func (s *ServerTestSuite) TestSetupServerNoH2() {
	serv := server.StaticServer{DisableH2: true}
	httpServer := server.GetServer(serv)
	s.NotNil(httpServer.TLSNextProto)
}
