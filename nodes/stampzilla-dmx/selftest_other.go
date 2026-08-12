//go:build !linux

package main

// maybeRunSelftest is only implemented on Linux (see selftest.go), matching
// openOpenDMXOutput. It always defers to normal node startup on other
// platforms.
func maybeRunSelftest(_ []string) (bool, int) {
	return false, 0
}
