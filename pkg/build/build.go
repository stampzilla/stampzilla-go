package build

import "fmt"

// Version is used by CI/CD system to set the version of the built binary.
var Version = "dev"

// BuildTime is the time of this build.
var BuildTime = ""

// Commit contains the SHA commit hash.
var Commit = ""

// String returns the version info as a string.
func String() string {
	return fmt.Sprintf(`Version: "%s", BuildTime: "%s", Commit: "%s" `, Version, BuildTime, Commit)
}

func JSON() string {
	return fmt.Sprintf(`{"Version": "%s", "BuildTime": "%s", "Commit": "%s"} `, Version, BuildTime, Commit)
}
