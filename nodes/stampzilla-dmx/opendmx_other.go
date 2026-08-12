//go:build !linux

package main

import "fmt"

// openOpenDMXOutput is only implemented on Linux (see opendmx_linux.go),
// which is the only platform stampzilla nodes are built/shipped for
// (cmd/build/build.go). This stub keeps `go build ./...`/`go vet ./...`
// green for contributors on other platforms.
func openOpenDMXOutput(_ string, _ breakMode, _ deMode) (dmxOutput, error) {
	return nil, fmt.Errorf("dmx: open-dmx output is only supported on linux")
}
