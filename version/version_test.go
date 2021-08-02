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

// App reports the application name and version.
func (s *VersionTestSuite) TestApp() {
	s.Equal("h2static", version.App.Name)
	s.NotEqual("", version.App.Version)
}

// String returns application name and version as a string.
func (s *VersionTestSuite) TestString() {
	v := version.Version{
		Name:    "myapp",
		Version: "1.2.3",
	}
	s.Equal("myapp 1.2.3", v.String())
}

// Identifier returns slash-separated name and version.
func (s *VersionTestSuite) TestIdentifier() {
	v := version.Version{
		Name:    "myapp",
		Version: "1.2.3",
	}
	s.Equal("myapp/1.2.3", v.Identifier())
}
