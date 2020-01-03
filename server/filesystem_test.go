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

func TestFileSystem(t *testing.T) {
	suite.Run(t, new(FileSystemTestSuite))
}

type FileSystemTestSuite struct {
	suite.Suite

	tempdir string
	dir     http.Dir
}

func (s *FileSystemTestSuite) SetupTest() {
	tempdir, err := ioutil.TempDir("", "fs")
	if err != nil {
		panic(err)
	}

	s.tempdir = tempdir
	s.dir = http.Dir(tempdir)

}

func (s *FileSystemTestSuite) TearDownTest() {
	os.RemoveAll(s.tempdir)
}

func (s *FileSystemTestSuite) makeFile(name, content string) {
	fd := filepath.Join(s.tempdir, name)
	err := ioutil.WriteFile(fd, []byte(content), 0644)
	s.Nil(err)
}

func (s *FileSystemTestSuite) fileList(file http.File) (names []string) {
	fileInfos, err := file.Readdir(-1)
	for _, fileInfo := range fileInfos {
		names = append(names, fileInfo.Name())
	}
	s.Nil(err)
	return
}

// The file with .html suffix is returned if present
func (s *FileSystemTestSuite) TestLookupWithHTMLSuffix() {
	s.makeFile("test.html", "foo")
	fs := server.FileSystem{FileSystem: s.dir, ResolveHTML: true}
	file, err := fs.Open("/test")
	s.Nil(err)
	content, err := ioutil.ReadAll(file)
	s.Nil(err)
	s.Equal(content, []byte("foo"))
}

// The file with .htm suffix is returned if present
func (s *FileSystemTestSuite) TestLookupWithHTMSuffix() {
	s.makeFile("test.htm", "foo")
	fs := server.FileSystem{FileSystem: s.dir, ResolveHTML: true}
	file, err := fs.Open("/test")
	s.Nil(err)
	content, err := ioutil.ReadAll(file)
	s.Nil(err)
	s.Equal(content, []byte("foo"))
}

// The file with .html suffix is preferred over the .htm one if both are present
func (s *FileSystemTestSuite) TestLookupWithHTMLSuffixPreferred() {
	s.makeFile("test.html", "with html")
	s.makeFile("test.htm", "with htm")
	fs := server.FileSystem{FileSystem: s.dir, ResolveHTML: true}
	file, err := fs.Open("/test")
	s.Nil(err)
	content, err := ioutil.ReadAll(file)
	s.Nil(err)
	s.Equal(content, []byte("with html"))
}

// The suffix is not added if the name with suffix is a directory
func (s *FileSystemTestSuite) TestLookupWithSuffixNotIfDirectory() {
	err := os.Mkdir(filepath.Join(s.tempdir, "test.html"), 0755)
	s.Nil(err)
	err = os.Mkdir(filepath.Join(s.tempdir, "test.htm"), 0755)
	s.Nil(err)
	fs := server.FileSystem{FileSystem: s.dir, ResolveHTML: true}
	file, err := fs.Open("/test")
	s.IsType(err, &os.PathError{})
	s.Nil(file)
}

// Files with not .htm(l) suffix are not looked up if the option is disabled
func (s *FileSystemTestSuite) TestNoLookupWithHTMLSuffix() {
	s.makeFile("test.html", "")
	s.makeFile("test.htm", "")
	fs := server.FileSystem{FileSystem: s.dir, ResolveHTML: false}
	file, err := fs.Open("/test")
	s.IsType(err, &os.PathError{})
	s.Nil(file)
}

// Files starting with a dot can be hidden.
func (s *FileSystemTestSuite) TestHideDotFiles() {
	s.makeFile(".foo", "")
	fs := server.FileSystem{FileSystem: s.dir, HideDotFiles: true}
	file, err := fs.Open("/.foo")
	s.IsType(err, os.ErrNotExist)
	s.Nil(file)
}

// Files starting with a dot can be shown.
func (s *FileSystemTestSuite) TestShowDotFiles() {
	s.makeFile(".foo", "hidden foo")
	fs := server.FileSystem{FileSystem: s.dir, HideDotFiles: false}
	file, err := fs.Open("/.foo")
	s.Nil(err)
	content, err := ioutil.ReadAll(file)
	s.Nil(err)
	s.Equal(content, []byte("hidden foo"))
}

// Files starting with a dot can be hidden from listing.
func (s *FileSystemTestSuite) TestHideDotFilesListing() {
	s.makeFile(".foo", "")
	s.makeFile(".bar", "")
	s.makeFile("baz", "")
	s.makeFile("bza", "")
	fs := server.FileSystem{FileSystem: s.dir, HideDotFiles: true}
	file, err := fs.Open("/")
	s.Nil(err)
	s.Equal(s.fileList(file), []string{"baz", "bza"})
}

// Files starting with a dot can be included in listing.
func (s *FileSystemTestSuite) TestShowDotFilesListing() {
	s.makeFile(".foo", "")
	s.makeFile("bar", "")
	fs := server.FileSystem{FileSystem: s.dir, HideDotFiles: false}
	file, err := fs.Open("/")
	s.Nil(err)
	s.Equal(s.fileList(file), []string{".foo", "bar"})
}
