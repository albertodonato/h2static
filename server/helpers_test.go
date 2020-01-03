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

// WriteFile creates a filw with the specified content, returning the absolute
// path.
func (s *TempDirTestSuite) WriteFile(name, content string) string {
	absPath := filepath.Join(s.TempDir, name)
	err := ioutil.WriteFile(absPath, []byte(content), 0644)
	s.Nil(err)
	return absPath
}
