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
	TempDirTestSuite

	dir http.Dir
}

func (s *FileSystemTestSuite) SetupTest() {
	s.TempDirTestSuite.SetupTest()
	s.dir = http.Dir(s.TempDir)
}

func (s *FileSystemTestSuite) fileList(file http.File) (names []string) {
	fileInfos, err := file.Readdir(-1)
	for _, fileInfo := range fileInfos {
		names = append(names, fileInfo.Name())
	}
	s.Nil(err)
	return
}

// NewFileSystem returns a new filesystem.
func (s *FileSystemTestSuite) TestNewFileSystem() {
	fs := server.NewFileSystem("/some/dir", true, true)
	s.Equal(http.Dir("/some/dir"), fs.FileSystem)
	s.Equal("/some/dir", fs.Root)
	s.True(fs.ResolveHTML)
	s.True(fs.HideDotFiles)
}

// The file with .html suffix is returned if present
func (s *FileSystemTestSuite) TestLookupWithHTMLSuffix() {
	s.WriteFile("test.html", "foo")
	fs := server.FileSystem{FileSystem: s.dir, ResolveHTML: true}
	file, err := fs.Open("/test")
	s.Nil(err)
	content, err := ioutil.ReadAll(file)
	s.Nil(err)
	s.Equal([]byte("foo"), content)
}

// The file with .htm suffix is returned if present
func (s *FileSystemTestSuite) TestLookupWithHTMSuffix() {
	s.WriteFile("test.htm", "foo")
	fs := server.FileSystem{FileSystem: s.dir, ResolveHTML: true}
	file, err := fs.Open("/test")
	s.Nil(err)
	content, err := ioutil.ReadAll(file)
	s.Nil(err)
	s.Equal([]byte("foo"), content)
}

// The file with .html suffix is preferred over the .htm one if both are present
func (s *FileSystemTestSuite) TestLookupWithHTMLSuffixPreferred() {
	s.WriteFile("test.html", "with html")
	s.WriteFile("test.htm", "with htm")
	fs := server.FileSystem{FileSystem: s.dir, ResolveHTML: true}
	file, err := fs.Open("/test")
	s.Nil(err)
	content, err := ioutil.ReadAll(file)
	s.Nil(err)
	s.Equal([]byte("with html"), content)
}

// The suffix is not added if the name with suffix is a directory
func (s *FileSystemTestSuite) TestLookupWithSuffixNotIfDirectory() {
	err := os.Mkdir(filepath.Join(s.TempDir, "test.html"), 0755)
	s.Nil(err)
	err = os.Mkdir(filepath.Join(s.TempDir, "test.htm"), 0755)
	s.Nil(err)
	fs := server.FileSystem{FileSystem: s.dir, ResolveHTML: true}
	file, err := fs.Open("/test")
	s.IsType(&os.PathError{}, err)
	s.Nil(file)
}

// Files with not .htm(l) suffix are not looked up if the option is disabled
func (s *FileSystemTestSuite) TestNoLookupWithHTMLSuffix() {
	s.WriteFile("test.html", "")
	s.WriteFile("test.htm", "")
	fs := server.FileSystem{FileSystem: s.dir, ResolveHTML: false}
	file, err := fs.Open("/test")
	s.IsType(&os.PathError{}, err)
	s.Nil(file)
}

// Files starting with a dot can be hidden.
func (s *FileSystemTestSuite) TestHideDotFiles() {
	s.WriteFile(".foo", "")
	fs := server.FileSystem{FileSystem: s.dir, HideDotFiles: true}
	file, err := fs.Open("/.foo")
	s.IsType(os.ErrNotExist, err)
	s.Nil(file)
}

// Files starting with a dot can be shown.
func (s *FileSystemTestSuite) TestShowDotFiles() {
	s.WriteFile(".foo", "hidden foo")
	fs := server.FileSystem{FileSystem: s.dir, HideDotFiles: false}
	file, err := fs.Open("/.foo")
	s.Nil(err)
	content, err := ioutil.ReadAll(file)
	s.Nil(err)
	s.Equal([]byte("hidden foo"), content)
}

// Files starting with a dot can be hidden from listing.
func (s *FileSystemTestSuite) TestHideDotFilesListing() {
	s.WriteFile(".foo", "")
	s.WriteFile(".bar", "")
	s.WriteFile("baz", "")
	s.WriteFile("bza", "")
	fs := server.FileSystem{FileSystem: s.dir, HideDotFiles: true}
	file, err := fs.Open("/")
	s.Nil(err)
	s.Equal([]string{"baz", "bza"}, s.fileList(file))
}

// Files starting with a dot can be included in listing.
func (s *FileSystemTestSuite) TestShowDotFilesListing() {
	s.WriteFile(".foo", "")
	s.WriteFile("bar", "")
	fs := server.FileSystem{FileSystem: s.dir, HideDotFiles: false}
	file, err := fs.Open("/")
	s.Nil(err)
	s.Equal([]string{".foo", "bar"}, s.fileList(file))
}

// OpenFile returns a File if it's not a directory.
func (s *FileSystemTestSuite) TestOpenFileForFile() {
	s.WriteFile("foo", "bar")
	fs := server.FileSystem{FileSystem: s.dir}
	file, err := fs.OpenFile("/foo")
	s.Nil(err)
	content, err := ioutil.ReadAll(file)
	s.Nil(err)
	s.Equal([]byte("bar"), content)
}

// OpenFile errors if the File is a directory.
func (s *FileSystemTestSuite) TestOpenFileForDirectory() {
	fs := server.FileSystem{FileSystem: s.dir}
	file, err := fs.OpenFile("/")
	s.Nil(file)
	s.IsType(os.ErrNotExist, err)
}
