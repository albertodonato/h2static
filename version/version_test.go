package version_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/albertodonato/h2static/version"
)

func TestVersion(t *testing.T) {
	suite.Run(t, new(VersionTestSuite))
}

type VersionTestSuite struct {
	suite.Suite
}

func (s *VersionTestSuite) TestApp() {
	s.Equal("h2static", version.App.Name)
	s.NotEqual("", version.App.Version)
}
