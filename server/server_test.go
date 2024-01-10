package server_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/albertodonato/h2static/server"
	"github.com/albertodonato/h2static/testhelpers"
)

var nonExistentPath string

func init() {
	if runtime.GOOS == "windows" {
		nonExistentPath = `C:\not\here`
	} else {
		nonExistentPath = "/not/here"
	}
}

func TestStaticServerConfig(t *testing.T) {
	suite.Run(t, new(StaticServerConfigTestSuite))
}

type StaticServerConfigTestSuite struct {
	testhelpers.TempDirTestSuite
}

// If the specified Dir doesn't exist, an error is returned.
func (s *StaticServerConfigTestSuite) TestConfigValidateDirNotExists() {
	config := server.StaticServerConfig{Dir: nonExistentPath}
	err := config.Validate()
	s.NotNil(err)
	s.Contains(err.Error(), nonExistentPath)
}

// If the specified Dir exists but is not a directory, an error is returned.
func (s *StaticServerConfigTestSuite) TestConfigValidateDirNotDir() {
	path := s.WriteFile("foo", "bar")
	config := server.StaticServerConfig{Dir: path}
	err := config.Validate()
	s.NotNil(err)
	s.Equal(fmt.Sprintf("not a directory: %s", path), err.Error())
}

// If the CSS file doesn't exist, an error is returned.
func (s *StaticServerConfigTestSuite) TestConfigValidateCSSFileNotExist() {
	config := server.StaticServerConfig{
		Dir: s.TempDir,
		CSS: nonExistentPath,
	}
	err := config.Validate()
	s.NotNil(err)
	s.Contains(err.Error(), nonExistentPath)
}

// If the TLS certificate file doesn't exist, an error is returned.
func (s *StaticServerConfigTestSuite) TestConfigValidateTLSCertFileNotExists() {
	tlsKey := s.WriteFile("foo", "bar")
	config := server.StaticServerConfig{
		Dir:     s.TempDir,
		TLSCert: nonExistentPath,
		TLSKey:  tlsKey,
	}
	err := config.Validate()
	s.NotNil(err)
	s.Contains(err.Error(), nonExistentPath)
}

// If the TLS key file doesn't exist, an error is returned.
func (s *StaticServerConfigTestSuite) TestConfigValidateTLSKeyFileNotExists() {
	tlsCert := s.WriteFile("foo", "bar")
	config := server.StaticServerConfig{
		Dir:     s.TempDir,
		TLSCert: tlsCert,
		TLSKey:  nonExistentPath,
	}
	err := config.Validate()
	s.NotNil(err)
	s.Contains(err.Error(), nonExistentPath)
}

// If the passwords file doesn't exist, an error is returned.
func (s *StaticServerConfigTestSuite) TestConfigValidatePasswordFileNotExists() {
	config := server.StaticServerConfig{
		Dir:          s.TempDir,
		PasswordFile: nonExistentPath,
	}
	err := config.Validate()
	s.NotNil(err)
	s.Contains(err.Error(), nonExistentPath)
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
	var config server.StaticServerConfig
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
	var serv server.StaticServer
	s.False(serv.Config.IsHTTPS())
	s.Equal(serv.Scheme(), "http")
}

// GetServer returns static file handlers for a path.
func (s *StaticServerTestSuite) TestGetServerDefaultStaticServe() {
	content := "some content"
	s.WriteFile("test.txt", content)

	serv, err := server.NewStaticServer(server.StaticServerConfig{
		Dir: s.TempDir,
	})
	s.Nil(err)
	httpServer, err := server.GetServer(serv)
	s.Nil(err)
	s.Nil(httpServer.TLSNextProto)
	s.Equal("", httpServer.Addr)

	r := httptest.NewRequest("GET", "/test.txt", nil)
	w := httptest.NewRecorder()
	httpServer.Handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusOK, response.StatusCode)
	s.Equal(content, w.Body.String())
}

// GetServer returns a configured http.Server which allows outside symlink access
func (s *StaticServerTestSuite) TestSetupServerSpecifyAllowOutsideSymlinks() {
	content := "some content"
	s.WriteFile("outside.txt", content)
	subdir := s.Mkdir("sub")
	s.Symlink("../outside.txt", "sub/test.txt")
	serv, err := server.NewStaticServer(server.StaticServerConfig{
		Dir:                  subdir,
		AllowOutsideSymlinks: true,
	})
	s.Nil(err)
	httpServer, err := server.GetServer(serv)
	s.Nil(err)

	r := httptest.NewRequest("GET", "/test.txt", nil)
	w := httptest.NewRecorder()
	httpServer.Handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusOK, response.StatusCode)
	s.Equal(content, w.Body.String())
}

// GetServer returns a configured http.Server without HTTP/2
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

// GetServer returns a configured http.Server with Basic-Auth
func (s *StaticServerTestSuite) TestSetupServerBasicAuth() {
	absPath := s.WriteFile("basic-auth", "")
	serv, err := server.NewStaticServer(server.StaticServerConfig{
		Dir:          s.TempDir,
		PasswordFile: absPath,
	})
	s.Nil(err)
	httpServer, err := server.GetServer(serv)
	s.Nil(err)
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	httpServer.Handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusUnauthorized, response.StatusCode)
	s.Equal(
		`Basic realm="h2static", charset="UTF-8"`,
		response.Header.Get("WWW-Authenticate"))
}

func TestServeResources(t *testing.T) {
	suite.Run(t, new(ServeResourcesTestSuite))
}

type ServeResourcesTestSuite struct {
	testhelpers.TempDirTestSuite

	server *server.StaticServer
}

func (s *ServeResourcesTestSuite) SetupTest() {
	s.TempDirTestSuite.SetupTest()
	cssPath := s.WriteFile("style.css", "")
	server, err := server.NewStaticServer(server.StaticServerConfig{
		Dir: s.TempDir,
		CSS: cssPath,
	})
	s.server = server
	s.Nil(err)
}

// A Custom CSS file can be specified
func (s *ServeResourcesTestSuite) TestCSSFile() {
	cssContent := "body {background: red;}"
	s.WriteFile("style.css", cssContent)
	httpServer, err := server.GetServer(s.server)
	s.Nil(err)
	r := httptest.NewRequest("GET", server.CSSAsset, nil)
	w := httptest.NewRecorder()
	httpServer.Handler.ServeHTTP(w, r)
	response := w.Result()
	s.Equal(http.StatusOK, response.StatusCode)
	s.Equal(cssContent, w.Body.String())
}
