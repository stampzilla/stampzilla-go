package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// rgb holds a color as three components in the 0.0-1.0 range.
type rgb struct {
	R, G, B float64
}

// parseHexColor parses a "#rgb" or "#rrggbb" (the leading "#" is optional)
// string into an rgb with 0.0-1.0 components.
func parseHexColor(s string) (rgb, error) {
	orig := s
	s = strings.TrimPrefix(s, "#")

	var rs, gs, bs string
	switch len(s) {
	case 3:
		rs, gs, bs = s[0:1]+s[0:1], s[1:2]+s[1:2], s[2:3]+s[2:3]
	case 6:
		rs, gs, bs = s[0:2], s[2:4], s[4:6]
	default:
		return rgb{}, fmt.Errorf("dmx: invalid color %q: expected #rgb or #rrggbb", orig)
	}

	r, err := strconv.ParseUint(rs, 16, 8)
	if err != nil {
		return rgb{}, fmt.Errorf("dmx: invalid color %q: %w", orig, err)
	}
	g, err := strconv.ParseUint(gs, 16, 8)
	if err != nil {
		return rgb{}, fmt.Errorf("dmx: invalid color %q: %w", orig, err)
	}
	b, err := strconv.ParseUint(bs, 16, 8)
	if err != nil {
		return rgb{}, fmt.Errorf("dmx: invalid color %q: %w", orig, err)
	}

	return rgb{R: float64(r) / 255, G: float64(g) / 255, B: float64(b) / 255}, nil
}

// clamp01 clamps v to the 0.0-1.0 range.
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// scaleByte converts a 0.0-1.0 value (clamped) to a DMX channel byte (0-255).
func scaleByte(v float64) byte {
	return byte(math.Round(clamp01(v) * 255))
}

// hsvToRGB converts a hue/saturation/value color (h wraps around 0.0-1.0,
// s and v are 0.0-1.0) to rgb.
func hsvToRGB(h, s, v float64) rgb {
	h -= math.Floor(h)
	if s <= 0 {
		return rgb{v, v, v}
	}

	h6 := h * 6
	i := int(math.Floor(h6))
	f := h6 - float64(i)
	p := v * (1 - s)
	q := v * (1 - s*f)
	t := v * (1 - s*(1-f))

	switch i % 6 {
	case 0:
		return rgb{v, t, p}
	case 1:
		return rgb{q, v, p}
	case 2:
		return rgb{p, v, t}
	case 3:
		return rgb{p, q, v}
	case 4:
		return rgb{t, p, v}
	default:
		return rgb{v, p, q}
	}
}
