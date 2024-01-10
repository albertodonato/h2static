package testhelpers

import (
	"os"
	"path/filepath"

	"github.com/stretchr/testify/suite"
)

// TempDirTestSuite is a test suite which creates and cleanups a temporary
// directory.
type TempDirTestSuite struct {
	suite.Suite

	TempDir string
}

// SetupTest sets up a temporary dir.
func (s *TempDirTestSuite) SetupTest() {
	tempdir, err := os.MkdirTemp("", "testdir")
	s.Nil(err)
	s.TempDir = tempdir
}

// TearDownTest cleans up the temporary dir.
func (s *TempDirTestSuite) TearDownTest() {
	os.RemoveAll(s.TempDir)
}

// WriteFile creates a file with the specified content, returning the absolute
// path.
func (s *TempDirTestSuite) WriteFile(name, content string) string {
	path := s.absPath(name)
	err := os.WriteFile(path, []byte(content), 0644)
	s.Nil(err)
	return path
}

// Stat returns a FileInfo for the path.
func (s *TempDirTestSuite) Stat(name string) os.FileInfo {
	path := s.absPath(name)
	fileInfo, err := os.Stat(path)
	s.Nil(err)
	return fileInfo
}

// Mkdir creates a directory, returning the absolute path.
func (s *TempDirTestSuite) Mkdir(name string) string {
	path := s.absPath(name)
	err := os.Mkdir(path, 0755)
	s.Nil(err)
	return path
}

// Symlink creates a symbolic link to oldname returning the absolute path of
// the new name. Both paths are relative to the tempdir path.
func (s *TempDirTestSuite) Symlink(oldname, newname string) string {
	newPath := s.absPath(newname)
	err := os.Symlink(oldname, newPath)
	s.Nil(err)
	return newPath
}

// RemoveAll removes the path and everything under it.
func (s *TempDirTestSuite) RemoveAll(name string) {
	path := s.absPath(name)
	err := os.RemoveAll(path)
	s.Nil(err)
}

func (s *TempDirTestSuite) absPath(path string) string {
	return filepath.Join(s.TempDir, path)
}
