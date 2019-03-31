package main

import (
	"flag"
	"log"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestH2Static(t *testing.T) {
	suite.Run(t, new(H2StaticTestSuite))
}

// A writer that collects the content
type collectWriter struct {
	content []byte
}

func (l *collectWriter) Write(p []byte) (int, error) {
	l.content = append(l.content, p...)
	return len(p), nil
}

func (l collectWriter) Output() string {
	return string(l.content)
}

type H2StaticTestSuite struct {
	suite.Suite

	logger *log.Logger
}

func (s *H2StaticTestSuite) SetupSuite() {
	s.logger = log.New(&collectWriter{}, "", 0)
}

// parseFlags parses commandline options
func (s *H2StaticTestSuite) TestParseFlags() {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	flags, err := parseFlags(
		flagSet,
		[]string{
			"-addr", ":9090", "-dir", "somedir", "-disable-h2",
			"-log", "-tls-cert", "crt", "-tls-key", "key"})
	s.Nil(err)
	s.Equal(flags.Addr, ":9090")
	s.Equal(flags.Dir, "somedir")
	s.True(flags.DisableH2)
	s.True(flags.Log)
	s.Equal(flags.TLSCert, "crt")
	s.Equal(flags.TLSKey, "key")
}

// parseFlags prints help
func (s *H2StaticTestSuite) TestParseFlagsHelp() {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	writer := &collectWriter{}
	flagSet.SetOutput(writer)
	_, err := parseFlags(flagSet, []string{"-h"})
	s.Equal(err, flag.ErrHelp)
	s.Contains(
		writer.Output(), "Tiny static web server with TLS and HTTP/2 support.")
}

// enableTLS returns true if certs are set
func (s *H2StaticTestSuite) TestEnableTLSTrue() {
	flags := cmdFlags{
		TLSCert: "cert",
		TLSKey:  "secret",
	}
	s.True(enableTLS(flags))
}

// enableTLS returns false if certs are not set
func (s *H2StaticTestSuite) TestEnableTLSFalse() {
	s.False(enableTLS(cmdFlags{}))
}

// setupServer returns a configured http.Server
func (s *H2StaticTestSuite) TestSetupServerDefault() {
	server := setupServer(cmdFlags{})
	s.Equal(server.Addr, "")
	s.IsType(server.Handler, http.FileServer(http.Dir(".")))
	s.Nil(server.TLSNextProto)
}

// setupServer returns a configured http.Server with the specified dir
func (s *H2StaticTestSuite) TestSetupServerSpecifyDir() {
	server := setupServer(cmdFlags{Dir: "/some/dir"})
	s.Equal(server.Addr, "")
	s.IsType(server.Handler, http.FileServer(http.Dir("/some/dir")))
}

// setupServer returns a configured http.Server with logging
func (s *H2StaticTestSuite) TestSetupServerLog() {
	server := setupServer(cmdFlags{Log: true})
	s.IsType(server.Handler, &loggingHandler{})
}

// setupServer returns a configured http.Server without HTTP/2
func (s *H2StaticTestSuite) TestSetupServerNoH2() {
	server := setupServer(cmdFlags{DisableH2: true})
	s.NotNil(server.TLSNextProto)
}
