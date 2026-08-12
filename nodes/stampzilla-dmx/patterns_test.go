package main

import (
	"math"
	"testing"
	"time"
)

func TestPatternStatic(t *testing.T) {
	colors := []rgb{{1, 0, 0}}
	out := patterns["static"](frame{Colors: colors})
	if out.Intensity != 1 || out.Color != colors[0] {
		t.Errorf("patternStatic = %+v", out)
	}
}

func TestPatternOff(t *testing.T) {
	out := patterns["off"](frame{Colors: []rgb{{1, 1, 1}}})
	if out.Intensity != 0 {
		t.Errorf("patternOff intensity = %v, want 0", out.Intensity)
	}
}

func TestPatternChase(t *testing.T) {
	colors := []rgb{{1, 0, 0}, {0, 1, 0}}
	count := 3
	for step := 0; step < count*2; step++ {
		lit := -1
		for i := 0; i < count; i++ {
			out := patterns["chase"](frame{Index: i, Count: count, Step: step, Colors: colors})
			if out.Intensity > 0 {
				if lit != -1 {
					t.Fatalf("chase step=%d: more than one fixture lit (%d and %d)", step, lit, i)
				}
				lit = i
			}
		}
		if want := step % count; lit != want {
			t.Errorf("chase step=%d: lit index=%d, want %d", step, lit, want)
		}
	}
}

func TestPatternFill(t *testing.T) {
	colors := []rgb{{1, 1, 1}}
	count := 4

	litAt := func(step int) []bool {
		lit := make([]bool, count)
		for i := 0; i < count; i++ {
			out := patterns["fill"](frame{Index: i, Count: count, Step: step, Colors: colors})
			lit[i] = out.Intensity > 0
		}
		return lit
	}

	tests := []struct {
		step int
		want []bool
	}{
		{0, []bool{true, false, false, false}},  // filling: light 0 on
		{1, []bool{true, true, false, false}},   // filling: lights 0,1 on
		{2, []bool{true, true, true, false}},    // filling: lights 0,1,2 on
		{3, []bool{true, true, true, true}},     // filling: all on
		{4, []bool{true, true, true, false}},    // draining: light 3 (last lit) turns off first
		{5, []bool{true, true, false, false}},   // draining: light 2 turns off next
		{6, []bool{true, false, false, false}},  // draining: light 1 turns off next
		{7, []bool{false, false, false, false}}, // draining: light 0 turns off last, all dark
		{8, []bool{true, false, false, false}},  // cycle repeats
	}
	for _, tt := range tests {
		got := litAt(tt.step)
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("fill step=%d: lit=%v, want %v", tt.step, got, tt.want)
				break
			}
		}
	}
}

func TestPatternFillOnceForwardHoldsFull(t *testing.T) {
	colors := []rgb{{1, 1, 1}}
	count := 4

	litAt := func(step int) []bool {
		lit := make([]bool, count)
		for i := 0; i < count; i++ {
			out := patterns["fillonce"](frame{Index: i, Count: count, Step: step, Colors: colors})
			lit[i] = out.Intensity > 0
		}
		return lit
	}

	tests := []struct {
		step int
		want []bool
	}{
		{0, []bool{true, false, false, false}},
		{1, []bool{true, true, false, false}},
		{2, []bool{true, true, true, false}},
		{3, []bool{true, true, true, true}},
		{4, []bool{true, true, true, true}},   // holds full, does not drain
		{99, []bool{true, true, true, true}},  // holds full forever, no modulo wraparound
	}
	for _, tt := range tests {
		got := litAt(tt.step)
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("fillonce step=%d: lit=%v, want %v", tt.step, got, tt.want)
				break
			}
		}
	}
}

func TestPatternFillOnceClosingHoldsDark(t *testing.T) {
	colors := []rgb{{1, 1, 1}}
	count := 4

	litAt := func(step int) []bool {
		lit := make([]bool, count)
		for i := 0; i < count; i++ {
			out := patterns["fillonce"](frame{Index: i, Count: count, Step: step, Colors: colors, Closing: true})
			lit[i] = out.Intensity > 0
		}
		return lit
	}

	tests := []struct {
		step int
		want []bool
	}{
		{0, []bool{true, true, true, false}},   // draining: light 3 (last) turns off first
		{1, []bool{true, true, false, false}},
		{2, []bool{true, false, false, false}},
		{3, []bool{false, false, false, false}}, // all dark
		{4, []bool{false, false, false, false}}, // holds dark, does not re-fill
		{99, []bool{false, false, false, false}},
	}
	for _, tt := range tests {
		got := litAt(tt.step)
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("fillonce closing step=%d: lit=%v, want %v", tt.step, got, tt.want)
				break
			}
		}
	}
}

func TestPatternFillOnceEdgeCounts(t *testing.T) {
	if out := patterns["fillonce"](frame{Count: 0}); out.Intensity != 0 {
		t.Errorf("fillonce with Count=0 = %+v, want zero output", out)
	}
	colors := []rgb{{1, 1, 1}}
	if out := patterns["fillonce"](frame{Index: 0, Count: 1, Step: 0, Colors: colors}); out.Intensity == 0 {
		t.Error("fillonce with Count=1 forward should be lit")
	}
	if out := patterns["fillonce"](frame{Index: 0, Count: 1, Step: 0, Colors: colors, Closing: true}); out.Intensity != 0 {
		t.Error("fillonce with Count=1 closing should be dark")
	}
}

// TestPatternFillOnceNegativeStepIsClamped guards against the worst-case
// asymmetry a negative step (e.g. a future startedAt) could otherwise cause:
// unclamped, closing's formula (Index < Count-1-step) would evaluate to
// Index < Count for step=-1, lighting every fixture instead of leaving the
// animation at its legitimate starting point. Step is never intentionally
// negative in production, but this should never be reachable as "worse than
// step=0" regardless.
func TestPatternFillOnceNegativeStepIsClamped(t *testing.T) {
	colors := []rgb{{1, 1, 1}}
	count := 4
	for i := 0; i < count; i++ {
		negative := patterns["fillonce"](frame{Index: i, Count: count, Step: -1, Colors: colors, Closing: true})
		zero := patterns["fillonce"](frame{Index: i, Count: count, Step: 0, Colors: colors, Closing: true})
		if negative.Intensity != zero.Intensity {
			t.Errorf("fillonce closing: index %d at step=-1 = %v, want same as step=0 (%v)", i, negative.Intensity, zero.Intensity)
		}
	}
}

func TestRendersWhileOffPatternsAreRegistered(t *testing.T) {
	for name := range rendersWhileOff {
		if _, ok := patterns[name]; !ok {
			t.Errorf("rendersWhileOff contains %q, which is not a registered pattern", name)
		}
	}
}

func TestFillOnceLitCountMatchesPatternOutput(t *testing.T) {
	colors := []rgb{{1, 1, 1}}
	count := 6

	countLit := func(step int, closing bool) int {
		n := 0
		for i := 0; i < count; i++ {
			out := patterns["fillonce"](frame{Index: i, Count: count, Step: step, Colors: colors, Closing: closing})
			if out.Intensity > 0 {
				n++
			}
		}
		return n
	}

	for _, closing := range []bool{false, true} {
		for step := 0; step < count+2; step++ {
			elapsed := time.Duration(step) * time.Second
			got := fillOnceLitCount(closing, elapsed, time.Second, count)
			want := countLit(step, closing)
			if got != want {
				t.Errorf("fillOnceLitCount(closing=%v, step=%d) = %d, want %d (from pattern output)", closing, step, got, want)
			}
		}
	}
}

func TestPatternAlternateTogglesBetweenSteps(t *testing.T) {
	colors := []rgb{{1, 1, 1}}
	a := patterns["alternate"](frame{Index: 0, Step: 0, Colors: colors})
	b := patterns["alternate"](frame{Index: 0, Step: 1, Colors: colors})
	if (a.Intensity > 0) == (b.Intensity > 0) {
		t.Errorf("alternate did not toggle between steps: %+v vs %+v", a, b)
	}
}

func TestPatternPulse(t *testing.T) {
	colors := []rgb{{1, 1, 1}}
	start := patterns["pulse"](frame{Phase: 0, Colors: colors})
	if math.Abs(start.Intensity) > 1e-9 {
		t.Errorf("pulse at phase 0 = %v, want ~0", start.Intensity)
	}
	peak := patterns["pulse"](frame{Phase: 0.5, Colors: colors})
	if math.Abs(peak.Intensity-1) > 1e-9 {
		t.Errorf("pulse at phase 0.5 = %v, want ~1", peak.Intensity)
	}
}

func TestPatternWavePhaseOffset(t *testing.T) {
	colors := []rgb{{1, 1, 1}}
	a := patterns["wave"](frame{Index: 0, Count: 2, Phase: 0, Colors: colors})
	b := patterns["wave"](frame{Index: 1, Count: 2, Phase: 0, Colors: colors})
	if math.Abs(a.Intensity-b.Intensity) < 1e-9 {
		t.Errorf("wave: fixtures in a 2-fixture group should differ in phase, both = %v", a.Intensity)
	}
}

func TestPatternColorCycle(t *testing.T) {
	colors := []rgb{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}
	for step, want := range colors {
		out := patterns["colorcycle"](frame{Step: step, Colors: colors})
		if out.Color != want {
			t.Errorf("colorcycle step=%d = %+v, want %+v", step, out.Color, want)
		}
	}
}

func TestPatternRainbowAtPhaseZeroIsRed(t *testing.T) {
	colors := []rgb{{1, 1, 1}}
	out := patterns["rainbow"](frame{Phase: 0, Colors: colors})
	if math.Abs(out.Color.R-1) > 1e-9 || out.Color.G > 1e-9 || out.Color.B > 1e-9 {
		t.Errorf("rainbow at phase 0 = %+v, want red", out.Color)
	}
}

func TestPatternRandomIsDeterministic(t *testing.T) {
	colors := []rgb{{1, 1, 1}, {0, 0, 1}}
	a := patterns["random"](frame{Step: 5, Index: 2, Colors: colors})
	b := patterns["random"](frame{Step: 5, Index: 2, Colors: colors})
	if a != b {
		t.Errorf("random pattern is not deterministic: %+v vs %+v", a, b)
	}
	c := patterns["random"](frame{Step: 6, Index: 2, Colors: colors})
	if a == c {
		t.Errorf("random pattern did not change between steps")
	}
}
