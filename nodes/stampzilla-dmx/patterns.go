package main

import (
	"fmt"
	"hash/fnv"
	"math"
)

// frame is the input a pattern function receives once per rendered frame,
// per fixture in the group.
type frame struct {
	Index   int     // fixture's position within the group (0-based)
	Count   int     // number of fixtures in the group
	Step    int     // increments by one every Interval
	Phase   float64 // 0.0-1.0 position within the current Step
	Colors  []rgb   // the group's configured colors, never empty
	Reverse bool
}

// patternOutput is what a pattern function computes for one fixture.
type patternOutput struct {
	Intensity float64 // 0.0-1.0, multiplied by the group's brightness
	Color     rgb
}

type patternFunc func(frame) patternOutput

// patterns is the registry of all built-in animated patterns, selectable via
// a group's "pattern" config/state key.
var patterns = map[string]patternFunc{
	"off":        patternOff,
	"static":     patternStatic,
	"chase":      patternChase,
	"fill":       patternFill,
	"alternate":  patternAlternate,
	"pulse":      patternPulse,
	"wave":       patternWave,
	"colorcycle": patternColorCycle,
	"rainbow":    patternRainbow,
	"random":     patternRandom,
}

// patternOff keeps every fixture dark.
func patternOff(frame) patternOutput {
	return patternOutput{}
}

// patternStatic lights every fixture at full intensity with the group's
// first configured color.
func patternStatic(f frame) patternOutput {
	return patternOutput{Intensity: 1, Color: f.Colors[0]}
}

// patternChase lights a single fixture at a time, walking through the group
// one Step at a time. The color advances through Colors with each Step.
func patternChase(f frame) patternOutput {
	if f.Count == 0 {
		return patternOutput{}
	}
	active := f.Step % f.Count
	if f.Index != active {
		return patternOutput{}
	}
	return patternOutput{Intensity: 1, Color: f.Colors[f.Step%len(f.Colors)]}
}

// patternFill lights fixtures up one at a time in order (0, 1, 2, ...) until
// the whole group is on, then turns them off one at a time starting from the
// last one that was lit (Count-1, Count-2, ...) until the group is dark
// again, then repeats.
func patternFill(f frame) patternOutput {
	if f.Count == 0 {
		return patternOutput{}
	}

	cycle := 2 * f.Count
	step := f.Step % cycle

	var on bool
	if step < f.Count {
		on = f.Index <= step
	} else {
		drainStep := step - f.Count
		on = f.Index < f.Count-1-drainStep
	}
	if !on {
		return patternOutput{}
	}
	return patternOutput{Intensity: 1, Color: f.Colors[0]}
}

// patternAlternate lights every other fixture, flipping which half is lit
// each Step.
func patternAlternate(f frame) patternOutput {
	if (f.Index+f.Step)%2 != 0 {
		return patternOutput{}
	}
	return patternOutput{Intensity: 1, Color: f.Colors[f.Step%len(f.Colors)]}
}

// patternPulse breathes all fixtures in the group in and out together.
func patternPulse(f frame) patternOutput {
	intensity := (1 - math.Cos(2*math.Pi*f.Phase)) / 2
	return patternOutput{Intensity: intensity, Color: f.Colors[0]}
}

// patternWave is like patternPulse, but each fixture is offset in phase by
// its position in the group, producing a wave running across the fixtures.
func patternWave(f frame) patternOutput {
	offset := 0.0
	if f.Count > 0 {
		offset = float64(f.Index) / float64(f.Count)
	}
	phase := f.Phase + offset
	phase -= math.Floor(phase)
	intensity := (1 - math.Cos(2*math.Pi*phase)) / 2
	return patternOutput{Intensity: intensity, Color: f.Colors[0]}
}

// patternColorCycle lights every fixture at full intensity, stepping through
// Colors one Step at a time.
func patternColorCycle(f frame) patternOutput {
	return patternOutput{Intensity: 1, Color: f.Colors[f.Step%len(f.Colors)]}
}

// patternRainbow sweeps the full hue circle across the group.
func patternRainbow(f frame) patternOutput {
	offset := 0.0
	if f.Count > 0 {
		offset = float64(f.Index) / float64(f.Count)
	}
	hue := f.Phase + offset
	hue -= math.Floor(hue)
	return patternOutput{Intensity: 1, Color: hsvToRGB(hue, 1, 1)}
}

// patternRandom picks a deterministic pseudo-random intensity and color per
// fixture, changing every Step. It intentionally avoids math/rand so output
// is reproducible (and unit-testable) for a given Step/Index.
func patternRandom(f frame) patternOutput {
	h := fnv.New32a()
	fmt.Fprintf(h, "%d:%d", f.Step, f.Index)
	v := h.Sum32()

	intensity := float64(v%1000) / 999.0
	color := f.Colors[int(v>>16)%len(f.Colors)]

	return patternOutput{Intensity: intensity, Color: color}
}
