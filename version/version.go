// Package version defines constants for application version.
package version

// Version tracks the application version details.
type Version struct {
	Name    string
	Version string
}

// The application version.
var App = Version{
	Name:    "h2static",
	Version: "1.2.0",
}
