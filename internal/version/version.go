package version

import (
	"fmt"
	"runtime"
)

// Set via ldflags at build time.
var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func String() string {
	return fmt.Sprintf("vex %s (%s) built %s %s/%s", Version, Commit, Date, runtime.GOOS, runtime.GOARCH)
}

func Short() string {
	return Version
}
