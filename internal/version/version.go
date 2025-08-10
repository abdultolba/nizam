package version

// Version is set at build time with -ldflags
var version = "0.8.0-dev"

// Version returns the current version
func Version() string {
	return version
}
