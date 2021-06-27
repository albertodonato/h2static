package testhelpers

import (
	"bytes"
	"log"
	"os"

	"github.com/stretchr/testify/suite"
)

// LogCaptureSuite is a test suite which captures log output to a buffer.
type LogCaptureSuite struct {
	suite.Suite

	Logs bytes.Buffer
}

// SetupTest redirect log output to a buffer.
func (s *LogCaptureSuite) SetupTest() {
	log.SetOutput(&s.Logs)
}

// TearDownTest resets log output to stderr.
func (s *LogCaptureSuite) TearDownTest() {
	log.SetOutput(os.Stderr)
}
