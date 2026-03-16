package version

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

// Set via ldflags at build time.
var (
	Version = ""
	Commit  = ""
	Date    = ""
)

func init() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	if Version == "" && info.Main.Version != "" && info.Main.Version != "(devel)" {
		Version = info.Main.Version
	}

	if Commit == "" {
		for _, s := range info.Settings {
			switch s.Key {
			case "vcs.revision":
				if len(s.Value) > 7 {
					Commit = s.Value[:7]
				} else {
					Commit = s.Value
				}
			case "vcs.time":
				if Date == "" {
					Date = s.Value
				}
			}
		}
	}

	if Version == "" {
		Version = "dev"
	}
	if Commit == "" {
		Commit = "unknown"
	}
	if Date == "" {
		Date = "unknown"
	}
}

func String() string {
	return fmt.Sprintf("vex %s (%s) built %s %s/%s", Version, Commit, Date, runtime.GOOS, runtime.GOARCH)
}

func Short() string {
	return Version
}
