// Package version defines constants for application version.
package version

// Version tracks the application version details.
type Version struct {
	Name    string
	Version string
}

// App defines The application version.
var App = Version{
	Name:    "h2static",
	Version: "2.1.0",
}
