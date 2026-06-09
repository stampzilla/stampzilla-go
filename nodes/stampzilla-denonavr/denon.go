package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/stampzilla/stampzilla-go/v2/nodes/stampzilla-server/models/devices"
)

// volumeToMV encodes a 0.0–1.0 volume to a Denon MV command parameter string.
// volume 1.0 maps to MV80 (0 dB), 0.0 maps to MV00 (−80 dB).
// Half-dB steps are encoded as a three-digit parameter, e.g. MV805 = −39.5 dB.
func volumeToMV(v float64) string {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	// Quantise to the nearest 0.5 dB step in the range [0, 80].
	raw := math.Round(v*80*2) / 2
	whole := int(raw)
	if raw == float64(whole) {
		return fmt.Sprintf("MV%02d", whole)
	}
	// Half-step: encode floor as two digits followed by "5".
	return fmt.Sprintf("MV%02d5", whole)
}

// mvToVolume decodes a Denon MV parameter string (e.g. "80", "805", "395") to a
// 0.0–1.0 volume. Returns −1 if the string cannot be parsed.
func mvToVolume(param string) float64 {
	switch len(param) {
	case 2:
		n, err := strconv.Atoi(param)
		if err != nil {
			return -1
		}
		return math.Min(float64(n)/80.0, 1.0)
	case 3:
		// e.g. "805" → 80.5 dB, "395" → 39.5 dB
		if param[2] != '5' {
			return -1
		}
		n, err := strconv.Atoi(param[:2])
		if err != nil {
			return -1
		}
		val := float64(n) + 0.5
		return math.Min(val/80.0, 1.0)
	default:
		return -1
	}
}

// parseEvent parses a single Denon event/response line (CR stripped) into a
// devices.State. Returns nil for lines that carry no state we track.
func parseEvent(line string) devices.State {
	switch {
	case line == "PWON":
		return devices.State{"on": true}
	case line == "PWSTANDBY":
		return devices.State{"on": false}
	case strings.HasPrefix(line, "MVMAX"):
		// MVMAX ## — the receiver echoes its maximum volume limit; ignore it.
		return nil
	case strings.HasPrefix(line, "MV"):
		v := mvToVolume(line[2:])
		if v < 0 {
			return nil
		}
		return devices.State{"volume": v}
	case strings.HasPrefix(line, "SI"):
		src := line[2:]
		if src == "" {
			return nil
		}
		return devices.State{"source": src}
	}
	return nil
}
