// Package version holds build information shared by all commands in hall-kit.
//
// The exported variables are injected at build time via -ldflags, e.g.:
//
//	go build -ldflags "-X github.com/QuietSugar/hall-kit/internal/version.Version=v1.0.0"
package version

// Build information. Populated at build time via -ldflags -X.
var (
	// Version is the semantic version of the build.
	Version = "dev"
	// Commit is the short hash of the git commit the build was made from.
	Commit = "none"
	// Date is the UTC build timestamp.
	Date = "unknown"
)

// String returns a human-readable, single-line version string.
func String() string {
	return "hall-kit " + Version + " (commit: " + Commit + ", built: " + Date + ")"
}
