package version

import "fmt"

var (
	// Version is overridden by release builds.
	Version = "v0.1.0"
	// Commit is overridden by release builds.
	Commit = "dev"
	// Date is overridden by release builds.
	Date = "unknown"
)

type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

func Info() BuildInfo {
	return BuildInfo{
		Version: Version,
		Commit:  Commit,
		Date:    Date,
	}
}

func (b BuildInfo) String() string {
	if b.Commit == "" || b.Commit == "dev" {
		return fmt.Sprintf("surveyctl %s", b.Version)
	}
	if b.Date == "" || b.Date == "unknown" {
		return fmt.Sprintf("surveyctl %s (%s)", b.Version, b.Commit)
	}
	return fmt.Sprintf("surveyctl %s (%s, built %s)", b.Version, b.Commit, b.Date)
}
