package version

import "fmt"

// These variables are set at build time via -ldflags by goreleaser.
// Default values are used during local development.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func String() string {
	return fmt.Sprintf("thermalik-bridge %s (commit %s, built %s)", Version, Commit, Date)
}
