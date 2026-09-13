package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the current semantic version of the CLI. Injected via -ldflags.
	Version = "0.1.0-dev"

	// GitCommit is the git commit sha at build time. Injected via -ldflags.
	GitCommit = "unknown"

	// BuildDate is the date of the build. Injected via -ldflags.
	BuildDate = "unknown"
)

// Info represents structured version information.
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// Get returns the structured version information.
func Get() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns human-readable version details.
func (i Info) String() string {
	return fmt.Sprintf("asactl version %s (commit: %s, date: %s, %s, %s)",
		i.Version, i.GitCommit, i.BuildDate, i.GoVersion, i.Platform)
}
