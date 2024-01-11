// Package version defines constants for application version.
package version

import (
	"fmt"
)

// Version tracks the application version details.
type Version struct {
	Name    string
	Version string
}

// String returns the version as a string.
func (v Version) String() string {
	return fmt.Sprintf("%s %s", v.Name, v.Version)
}

// Identifier returns the version intentifier.
func (v Version) Identifier() string {
	return fmt.Sprintf("%s/%s", v.Name, v.Version)
}

// App defines The application version.
var App = Version{
	Name:    "h2static",
	Version: "2.4.8",
}
