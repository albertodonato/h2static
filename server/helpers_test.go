package server_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/stretchr/testify/suite"
)

// TempDirTestSuite is a test suite which creates and cleanups a temporary
// directory
type TempDirTestSuite struct {
	suite.Suite

	TempDir string
}

func (s *TempDirTestSuite) SetupTest() {
	tempdir, err := ioutil.TempDir("", "testdir")
	if err != nil {
		panic(err)
	}

	s.TempDir = tempdir
}

func (s *TempDirTestSuite) TearDownTest() {
	os.RemoveAll(s.TempDir)
}

// WriteFile creates a file with the specified content, returning the absolute
// path.
func (s *TempDirTestSuite) WriteFile(name, content string) string {
	path := s.absPath(name)
	err := ioutil.WriteFile(path, []byte(content), 0644)
	s.Nil(err)
	return path
}

// Mkdir creates a directory, returning the absolute path.
func (s *TempDirTestSuite) Mkdir(name string) string {
	path := s.absPath(name)
	err := os.Mkdir(path, 0644)
	s.Nil(err)
	return path
}

func (s *TempDirTestSuite) absPath(path string) string {
	return filepath.Join(s.TempDir, path)
}
