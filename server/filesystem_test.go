package server_test

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/albertodonato/h2static/server"
)

func TestHTMLPageResolveFileSystem(t *testing.T) {
	suite.Run(t, new(HTMLPageResolveFileSystemTestSuite))
}

type HTMLPageResolveFileSystemTestSuite struct {
	suite.Suite

	tempdir string
	fs      server.HTMLPageResolveFileSystem
}

func (s *HTMLPageResolveFileSystemTestSuite) SetupTest() {
	tempdir, err := ioutil.TempDir("", "fs")
	if err != nil {
		panic(err)
	}

	s.tempdir = tempdir
	s.fs = server.HTMLPageResolveFileSystem{http.Dir(tempdir)}

}

func (s *HTMLPageResolveFileSystemTestSuite) TearDownTest() {
	os.RemoveAll(s.tempdir)
}

func (s *HTMLPageResolveFileSystemTestSuite) makeFile(name, content string) {
	fd := filepath.Join(s.tempdir, name)
	if err := ioutil.WriteFile(fd, []byte(content), 0644); err != nil {
		panic(err)
	}
}

// The file with .html suffix is returned if present
func (s *HTMLPageResolveFileSystemTestSuite) TestLookupWithHTMLSuffix() {
	s.makeFile("test.html", "foo")
	file, err := s.fs.Open("/test")
	s.Nil(err)
	content, err := ioutil.ReadAll(file)
	s.Nil(err)
	s.Equal(content, []byte("foo"))
}

// The file with .htm suffix is returned if present
func (s *HTMLPageResolveFileSystemTestSuite) TestLookupWithHTMSuffix() {
	s.makeFile("test.htm", "foo")
	file, err := s.fs.Open("/test")
	s.Nil(err)
	content, err := ioutil.ReadAll(file)
	s.Nil(err)
	s.Equal(content, []byte("foo"))
}

// The file with .html suffix is preferred over the .htm one if both are present
func (s *HTMLPageResolveFileSystemTestSuite) TestLookupWithHTMLSuffixPreferred() {
	s.makeFile("test.html", "with html")
	s.makeFile("test.htm", "with htm")
	file, err := s.fs.Open("/test")
	s.Nil(err)
	content, err := ioutil.ReadAll(file)
	s.Nil(err)
	s.Equal(content, []byte("with html"))
}

// The suffix is not added if the name with suffix is a directory
func (s *HTMLPageResolveFileSystemTestSuite) TestLookupWithSuffixNotIfDirectory() {
	err := os.Mkdir(filepath.Join(s.tempdir, "test.html"), 0755)
	s.Nil(err)
	err = os.Mkdir(filepath.Join(s.tempdir, "test.htm"), 0755)
	s.Nil(err)
	file, err := s.fs.Open("/test")
	s.IsType(err, &os.PathError{})
	s.Nil(file)
}
