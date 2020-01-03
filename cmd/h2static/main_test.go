package main_test

import (
	"flag"
	"log"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/albertodonato/h2static/cmd/h2static"
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

// NewStaticServerFromCmdline parses commandline options and returns a
// configured server.
func (s *H2StaticTestSuite) TestNewStaticServerFromCmdline() {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	server, err := main.NewStaticServerFromCmdline(
		flagSet,
		[]string{
			"-addr", ":9090", "-basic-auth", "passwords", "-dir", "somedir",
			"-disable-lookup-with-suffix", "-disable-h2", "-show-dotfiles",
			"-log", "-tls-cert", "crt", "-tls-key", "key"})
	s.Nil(err)
	s.Equal(":9090", server.Addr)
	s.Equal("passwords", server.PasswordFile)
	s.Equal("somedir", server.Dir)
	s.True(server.DisableH2)
	s.True(server.DisableLookupWithSuffix)
	s.True(server.ShowDotFiles)
	s.True(server.Log)
	s.Equal("crt", server.TLSCert)
	s.Equal("key", server.TLSKey)
}

// newStaticServerFromCmdline prints help text.
func (s *H2StaticTestSuite) TestParseFlagsHelp() {
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	writer := &collectWriter{}
	flagSet.SetOutput(writer)
	_, err := main.NewStaticServerFromCmdline(flagSet, []string{"-h"})
	s.Equal(flag.ErrHelp, err)
	s.Contains(
		writer.Output(), "Tiny static web server with TLS and HTTP/2 support.")
}
