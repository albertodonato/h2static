package main_test

import (
	"flag"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/albertodonato/h2static/cmd/h2static"
	"github.com/albertodonato/h2static/testhelpers"
)

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

func TestH2Static(t *testing.T) {
	suite.Run(t, new(H2StaticTestSuite))
}

type H2StaticTestSuite struct {
	testhelpers.TempDirTestSuite

	flagSet *flag.FlagSet
	writer  *collectWriter
}

func (s *H2StaticTestSuite) SetupTest() {
	s.TempDirTestSuite.SetupTest()

	s.flagSet = flag.NewFlagSet("test", flag.ContinueOnError)
	s.writer = &collectWriter{}
	s.flagSet.SetOutput(s.writer)
}

// NewStaticServerFromCmdline parses commandline options and returns a
// configured server.
func (s *H2StaticTestSuite) TestNewStaticServerFromCmdline() {
	passwdPath := s.WriteFile("passwords.txt", "some:password")
	certPath := s.WriteFile("crt.pem", "cert")
	keyPath := s.WriteFile("key.pem", "key")
	dirPath := s.Mkdir("dir")

	server, err := main.NewStaticServerFromCmdline(
		s.flagSet,
		[]string{
			"-addr", ":9090", "-allow-outside-symlinks", "-basic-auth", passwdPath,
			"-dir", dirPath, "-disable-lookup-with-suffix", "-disable-h2",
			"-show-dotfiles", "-log", "-tls-cert", certPath, "-tls-key", keyPath})
	s.Nil(err)
	s.Equal(":9090", server.Config.Addr)
	s.True(server.Config.AllowOutsideSymlinks)
	s.Equal(passwdPath, server.Config.PasswordFile)
	s.Equal(dirPath, server.Config.Dir)
	s.True(server.Config.DisableH2)
	s.True(server.Config.DisableLookupWithSuffix)
	s.True(server.Config.ShowDotFiles)
	s.True(server.Config.Log)
	s.Equal(certPath, server.Config.TLSCert)
	s.Equal(keyPath, server.Config.TLSKey)
}

// Config options are validated and error returned on invalid paths.
func (s *H2StaticTestSuite) TestValidateConfig() {
	fileName := filepath.Join("not", "here")
	server, err := main.NewStaticServerFromCmdline(
		s.flagSet, []string{"-dir", fileName})
	s.Nil(server)
	s.NotNil(err)
	s.Contains(err.Error(), fileName)
}

// newStaticServerFromCmdline prints help text.
func (s *H2StaticTestSuite) TestParseFlagsHelp() {
	_, err := main.NewStaticServerFromCmdline(s.flagSet, []string{"-h"})
	s.Equal(flag.ErrHelp, err)
	s.Contains(
		s.writer.Output(), "Tiny static web server with TLS and HTTP/2 support.")
}
