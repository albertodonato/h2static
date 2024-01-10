package server_test

import (
	"io"
	"os"
	"sort"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/albertodonato/h2static/server"
	"github.com/albertodonato/h2static/testhelpers"
)

func TestFileSystem(t *testing.T) {
	suite.Run(t, new(FileSystemTestSuite))
}

type FileSystemTestSuite struct {
	testhelpers.TempDirTestSuite

	fs server.FileSystem
}

func (s *FileSystemTestSuite) SetupTest() {
	s.TempDirTestSuite.SetupTest()
	s.fs = server.FileSystem{
		Root:         s.TempDir,
		ResolveHTML:  true,
		HideDotFiles: true,
	}
}

func (s *FileSystemTestSuite) readFile(file *server.File) string {
	f, err := os.Open(file.AbsPath())
	s.Nil(err)
	content, err := io.ReadAll(f)
	s.Nil(err)
	return string(content)
}

func (s *FileSystemTestSuite) fileList(file *server.File) (names []string) {
	files, err := file.Readdir()
	s.Nil(err)
	for _, file := range files {
		names = append(names, file.Info.Name())
	}
	sort.Strings(names)
	return
}

// The file with .html suffix is returned if present
func (s *FileSystemTestSuite) TestLookupWithHTMLSuffix() {
	s.WriteFile("test.html", "foo")
	s.fs.ResolveHTML = true
	file, err := s.fs.Open("/test")
	s.Nil(err)
	s.Equal("foo", s.readFile(file))
}

// The file with .htm suffix is returned if present
func (s *FileSystemTestSuite) TestLookupWithHTMSuffix() {
	s.WriteFile("test.htm", "foo")
	s.fs.ResolveHTML = true
	file, err := s.fs.Open("/test")
	s.Nil(err)
	s.Equal("foo", s.readFile(file))

}

// The file with .html suffix is preferred over the .htm one if both are present
func (s *FileSystemTestSuite) TestLookupWithHTMLSuffixPreferred() {
	s.WriteFile("test.html", "with html")
	s.WriteFile("test.htm", "with htm")
	s.fs.ResolveHTML = true
	file, err := s.fs.Open("/test")
	s.Nil(err)
	s.Equal("with html", s.readFile(file))
}

// The suffix is not added if the name with suffix is a directory
func (s *FileSystemTestSuite) TestLookupWithSuffixNotIfDirectory() {
	s.Mkdir("test.html")
	s.Mkdir("test.htm")
	s.fs.ResolveHTML = true
	file, err := s.fs.Open("/test")
	s.IsType(&os.PathError{}, err)
	s.Nil(file)
}

// Files with not .htm(l) suffix are not looked up if the option is disabled
func (s *FileSystemTestSuite) TestNoLookupWithHTMLSuffix() {
	s.WriteFile("test.html", "")
	s.WriteFile("test.htm", "")
	s.fs.ResolveHTML = false
	file, err := s.fs.Open("/test")
	s.IsType(&os.PathError{}, err)
	s.Nil(file)
}

// By default symlinks outside the filesystem root are not accessible.
func (s *FileSystemTestSuite) TestNoOutsideSymlink() {
	s.WriteFile("foo", "content")
	root := s.Mkdir("root")
	s.Symlink("../foo", "root/foo-link")
	fs := server.FileSystem{Root: root}
	file, err := fs.Open("/foo-link")
	s.IsType(os.ErrPermission, err)
	s.Nil(file)
}

// If enabled, symlinks outside the filesystem root are accessible.
func (s *FileSystemTestSuite) TestOutsideSymlinks() {
	s.WriteFile("foo", "content")
	root := s.Mkdir("root")
	s.Symlink("../foo", "root/foo-link")
	fs := server.FileSystem{Root: root, AllowOutsideSymlinks: true}
	file, err := fs.Open("/foo-link")
	s.Nil(err)
	s.Equal("content", s.readFile(file))
}

// Local symlinks are always accessible.
func (s *FileSystemTestSuite) TestLocalSymlinks() {
	root := s.Mkdir("root")
	s.WriteFile("root/foo", "content")
	s.Symlink("foo", "root/foo-link")
	fs := server.FileSystem{Root: root}
	file, err := fs.Open("/foo-link")
	s.Nil(err)
	s.Equal("content", s.readFile(file))
}

// Files starting with a dot can be hidden.
func (s *FileSystemTestSuite) TestHideDotFiles() {
	s.WriteFile(".foo", "")
	s.fs.HideDotFiles = true
	file, err := s.fs.Open("/.foo")
	s.IsType(os.ErrNotExist, err)
	s.Nil(file)
}

// Files starting with a dot can be shown.
func (s *FileSystemTestSuite) TestShowDotFiles() {
	s.WriteFile(".foo", "hidden foo")
	s.fs.HideDotFiles = false
	file, err := s.fs.Open("/.foo")
	s.Nil(err)
	s.Equal("hidden foo", s.readFile(file))
}

// Files starting with a dot can be hidden from listing.
func (s *FileSystemTestSuite) TestHideDotFilesListing() {
	s.WriteFile(".foo", "")
	s.WriteFile(".bar", "")
	s.WriteFile("baz", "")
	s.WriteFile("bza", "")
	s.fs.HideDotFiles = true
	file, err := s.fs.Open("/")
	s.Nil(err)
	s.Equal([]string{"baz", "bza"}, s.fileList(file))
}

// Files starting with a dot can be included in listing.
func (s *FileSystemTestSuite) TestShowDotFilesListing() {
	s.WriteFile(".foo", "")
	s.fs.HideDotFiles = false
	file, err := s.fs.Open("/")
	s.Nil(err)
	s.Equal([]string{".foo"}, s.fileList(file))
}

// Symlinks are reported as files or directories based on the target.
func (s *FileSystemTestSuite) TestListingWithSymlinks() {
	s.WriteFile("foo", "")
	s.Mkdir("bar")
	s.Symlink("foo", "new-foo")
	s.Symlink("bar", "new-bar")
	file, err := s.fs.Open("/")
	s.Nil(err)
	files, err := file.Readdir()
	s.Nil(err)
	type detail struct {
		name  string
		isdir bool
	}
	details := make([]detail, len(files))
	for i, file := range files {
		details[i] = detail{
			name:  file.Info.Name(),
			isdir: file.Info.IsDir(),
		}
	}
	sort.Slice(details, func(i, j int) bool {
		return details[i].name < details[j].name
	})
	s.Equal(
		[]detail{
			{name: "bar", isdir: true},
			{name: "foo", isdir: false},
			{name: "new-bar", isdir: true},
			{name: "new-foo", isdir: false},
		}, details)
}

// OpenFile returns a File if it's not a directory.
func (s *FileSystemTestSuite) TestOpenFileForFile() {
	s.WriteFile("foo", "bar")
	s.fs.HideDotFiles = true
	file, err := s.fs.OpenFile("/foo")
	s.Nil(err)
	s.Equal("bar", s.readFile(file))
}

// OpenFile errors if the File is a directory.
func (s *FileSystemTestSuite) TestOpenFileForDirectory() {
	file, err := s.fs.OpenFile("/")
	s.Nil(file)
	s.IsType(os.ErrNotExist, err)
}

func TestFile(t *testing.T) {
	suite.Run(t, new(FileTestSuite))
}

type FileTestSuite struct {
	testhelpers.TempDirTestSuite
}

func (s *FileTestSuite) TestNewFile() {
	f, err := server.NewFile(s.TempDir, true)
	s.Nil(err)
	s.Equal(s.TempDir, f.AbsPath())
}
