package main

import (
	"fmt"
	"hash/fnv"
	"math"
	"time"
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
	Closing bool // true while playing a rendersWhileOff pattern's off animation
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
	"scanner":    patternScanner,
	"fill":       patternFill,
	"fillonce":   patternFillOnce,
	"alternate":  patternAlternate,
	"pulse":      patternPulse,
	"wave":       patternWave,
	"colorcycle": patternColorCycle,
	"rainbow":    patternRainbow,
	"random":     patternRandom,
}

// rendersWhileOff marks patterns the engine keeps rendering (rather than
// skipping) for a while after a group's `on` goes false, so they can play a
// closing animation before actually going dark - see renderFrame's use of
// frame.Closing. Every other pattern is skipped the instant `on` is false,
// same as before this existed. Kept as an explicit opt-in set rather than
// inferring "closing" from whether the previous output was nonzero, which
// would need per-fixture history in the engine and break the "pattern is a
// pure function of frame" contract the backdated-startedAt tests rely on.
var rendersWhileOff = map[string]bool{
	"fillonce": true,
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

// patternScanner sweeps a single lit fixture back and forth across the
// group - a "Knight Rider"/KITT-style scanner: 0,1,2,...,Count-1,Count-2,
// ...,1,0,1,... Each end is visited once per bounce, not twice, so the
// light never pauses or double-hits at either end. Unlike patternChase, the
// color stays fixed at the group's first configured color as the light
// moves.
func patternScanner(f frame) patternOutput {
	if f.Count == 0 {
		return patternOutput{}
	}

	active := 0
	if f.Count > 1 {
		period := 2 * (f.Count - 1)
		pos := f.Step % period
		if pos >= f.Count {
			active = period - pos
		} else {
			active = pos
		}
	}

	if f.Index != active {
		return patternOutput{}
	}
	return patternOutput{Intensity: 1, Color: f.Colors[0]}
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

// patternFillOnce is like patternFill, but doesn't loop: while a group is on
// it fills fixtures in order (0, 1, 2, ...) once and then holds them lit;
// while switched off (f.Closing) it plays the same fill in reverse (last
// fixture first) once and then holds everything dark. See rendersWhileOff
// for how the engine keeps this pattern rendering after a group turns off,
// and fillOnceLitCount/fillOnceStartForLitCount for how on/off toggles carry
// visible progress across the flip instead of restarting from the extreme
// end.
func patternFillOnce(f frame) patternOutput {
	if f.Count == 0 {
		return patternOutput{}
	}

	step := f.Step
	if step < 0 {
		step = 0
	}
	if step > f.Count-1 {
		step = f.Count - 1
	}

	var on bool
	if !f.Closing {
		on = f.Index <= step
	} else {
		on = f.Index < f.Count-1-step
	}
	if !on {
		return patternOutput{}
	}
	return patternOutput{Intensity: 1, Color: f.Colors[0]}
}

// fillOnceLitCount returns how many fixtures a fillonce group currently has
// lit, given whether it's filling (closing=false) or draining (closing=true)
// and how long the current animation has been running. It mirrors
// patternFillOnce's own step math so the two can never disagree.
func fillOnceLitCount(closing bool, elapsed, interval time.Duration, count int) int {
	if interval <= 0 {
		interval = defaultInterval
	}
	step := int(elapsed / interval)
	if step < 0 {
		step = 0
	}
	if count > 0 && step > count-1 {
		step = count - 1
	}
	if !closing {
		return step + 1
	}
	lit := count - 1 - step
	if lit < 0 {
		lit = 0
	}
	return lit
}

// fillOnceStartForLitCount returns the startedAt that makes a fillonce
// group's animation, starting now, already show exactly lit fixtures in the
// given direction - used to carry visible progress across an on/off toggle
// instead of restarting the new direction from its extreme end (which would
// otherwise flash every fixture the old direction hadn't reached yet).
func fillOnceStartForLitCount(closing bool, lit, count int, interval time.Duration) time.Time {
	if interval <= 0 {
		interval = defaultInterval
	}
	var step int
	if !closing {
		step = lit - 1
	} else {
		step = count - 1 - lit
	}
	if step < 0 {
		step = 0
	}
	return time.Now().Add(-time.Duration(step) * interval)
}

// fillOnceSettledDark returns a startedAt far enough in the past that a
// fillonce group renders fully dark immediately - used when a group starts,
// or switches to fillonce, while already off, so it never flashes the tail
// of a drain animation it never actually played.
func fillOnceSettledDark(count int, interval time.Duration) time.Time {
	if interval <= 0 {
		interval = defaultInterval
	}
	return time.Now().Add(-time.Duration(count+1) * interval)
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
