package main

import (
	"flag"
	"log"
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
