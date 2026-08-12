//go:build !linux

package main

import (
	"flag"
	"fmt"
	"os"
)

// runSelftestCommand is only able to actually drive hardware on Linux (see
// selftest_linux.go), matching openOpenDMXOutput/opendmx_other.go. Argument
// parsing still runs here so `-h`/usage and validation errors behave
// identically across platforms; only execution is linux-only.
func runSelftestCommand(args []string) int {
	if _, err := parseSelftestArgs(args, os.Stderr); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}

	fmt.Fprintln(os.Stderr, "dmx: selftest is only supported on linux")
	return 1
}
