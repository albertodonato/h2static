//go:build linux || darwin

package server_test

import (
	"path/filepath"
	"syscall"

	"github.com/albertodonato/h2static/server"
)

// Special files are not included in listing.
func (s *FileSystemTestSuite) TestListingIgnoreSpecial() {
	fifoPath := filepath.Join(s.TempDir, "fifo")
	s.Nil(syscall.Mkfifo(fifoPath, 0644))
	file, err := s.fs.Open("/")
	s.Nil(err)
	files, err := file.Readdir()
	s.Nil(err)
	s.Equal([]*server.File{}, files)
}
