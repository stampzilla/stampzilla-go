package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
)

// duration wraps time.Duration so it can be configured as a JSON string such
// as "400ms" or "2s" (see time.ParseDuration).
type duration time.Duration

func (d *duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("dmx: invalid duration: %w", err)
	}
	if s == "" {
		*d = 0
		return nil
	}
	pd, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("dmx: invalid duration %q: %w", s, err)
	}
	*d = duration(pd)
	return nil
}

func (d duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// Profile describes a fixture's DMX channel layout: Channels[i] names the
// role of the channel at offset i from the fixture's address. Recognised
// roles are "dimmer", "red", "green", "blue", "white", "amber" and "uv";
// any other name (including "-") is treated as an opaque channel that is
// only ever driven by Static.
type Profile struct {
	Channels []string       `json:"channels"`
	Static   map[string]int `json:"static"`
}

// Fixture places a Profile at a DMX start address (1-512).
type Fixture struct {
	Profile string `json:"profile"`
	Address int    `json:"address"`
}

// Group is a set of fixtures driven together by a single animated pattern.
type Group struct {
	Name     string   `json:"name"`
	Fixtures []string `json:"fixtures"`
	Pattern  string   `json:"pattern"`
	Interval duration `json:"interval"`
	Colors   []string `json:"colors"`
	Reverse  bool     `json:"reverse"`
}

// Config is the free-form application config pushed from the stampzilla
// server (see config.example.json).
type Config struct {
	Port         string             `json:"port"`
	FPS          int                `json:"fps"`
	UniverseSize int                `json:"universeSize"`
	BreakMode    string             `json:"breakMode"`
	DEMode       string             `json:"deMode"`
	Profiles     map[string]Profile `json:"profiles"`
	Fixtures     map[string]Fixture `json:"fixtures"`
	Groups       map[string]Group   `json:"groups"`
}

const (
	defaultFPS          = 30
	maxFPS              = 44 // DMX-512's physical full-frame refresh limit
	defaultInterval     = 500 * time.Millisecond
	minUniverseSize     = 24
	dmxUniverseChannels = 512

	// microsPerDataByte is how long transmitting one DMX data byte takes at
	// 250,000 baud 8N2 (11 bits/byte). Used to sanity-check fps against
	// universeSize so renderFrame's write never overruns the frame ticker.
	microsPerDataByte = 44

	// minBreakOverhead is a conservative estimate of break+MAB transmission
	// time (180us break + 40us MAB, plus a safety margin for the baud
	// switches around it), added on top of data-byte time when sanity
	// checking fps against universeSize.
	minBreakOverhead = 500 * time.Microsecond
)

// builtinProfiles are always available and can be overridden by profiles
// declared in the config under the same name.
var builtinProfiles = map[string]Profile{
	"dimmer":      {Channels: []string{"dimmer"}},
	"rgb":         {Channels: []string{"red", "green", "blue"}},
	"rgbw":        {Channels: []string{"red", "green", "blue", "white"}},
	"rgba":        {Channels: []string{"red", "green", "blue", "amber"}},
	"rgb-dimmer":  {Channels: []string{"dimmer", "red", "green", "blue"}},
	"rgbw-dimmer": {Channels: []string{"dimmer", "red", "green", "blue", "white"}},
}

// resolvedProfile is a validated Profile ready to be used by the engine.
type resolvedProfile struct {
	channels []string     // offset -> role name
	static   map[int]byte // offset -> constant value, applied every frame
}

// roleOffset returns the channel offset of the given role, if the profile
// declares one.
func (p resolvedProfile) roleOffset(role string) (int, bool) {
	for i, c := range p.channels {
		if c == role {
			return i, true
		}
	}
	return 0, false
}

// resolvedFixture is a validated Fixture placed on the universe.
type resolvedFixture struct {
	address int // 1-based DMX start address
	profile resolvedProfile
}

// span returns the 0-based [start, end] byte range (inclusive) this fixture
// occupies in the universe buffer.
func (f resolvedFixture) span() (start, end int) {
	start = f.address - 1
	end = start + len(f.profile.channels) - 1
	return
}

// resolvedGroup is a validated Group ready to be used by the engine.
type resolvedGroup struct {
	name     string
	fixtures []string // fixture keys, in configured order
	pattern  string
	interval time.Duration
	colors   []rgb
	reverse  bool
}

// tickInterval returns g's pattern step interval, falling back to
// defaultInterval. Shared by renderFrame and onRequestStateChange so their
// step arithmetic (including fillonce's on/off toggle carry-over math) can
// never disagree.
func (g *resolvedGroup) tickInterval() time.Duration {
	if g.interval <= 0 {
		return defaultInterval
	}
	return g.interval
}

// resolvedConfig is a fully validated Config.
type resolvedConfig struct {
	port         string
	fps          int
	breakMode    breakMode
	deMode       deMode
	fixtures     map[string]resolvedFixture
	groups       map[string]*resolvedGroup
	universeSize int
}

// loadConfig parses and validates the raw JSON config received from the
// server.
func loadConfig(data json.RawMessage) (*resolvedConfig, error) {
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return resolveConfig(&c)
}

func resolveConfig(c *Config) (*resolvedConfig, error) {
	fps := c.FPS
	if fps == 0 {
		fps = defaultFPS
	}
	if fps > maxFPS {
		fps = maxFPS
	}
	if fps < 1 {
		return nil, fmt.Errorf("dmx: fps must be positive")
	}

	breakMode := breakMode(c.BreakMode)
	switch breakMode {
	case "":
		breakMode = breakModeBaud
	case breakModeBaud, breakModeIoctl:
	default:
		return nil, fmt.Errorf("dmx: unknown breakMode %q (want %q or %q)", c.BreakMode, breakModeBaud, breakModeIoctl)
	}

	deMode := deMode(c.DEMode)
	switch deMode {
	case "":
		deMode = deModeNone
	case deModeNone, deModeAssert, deModeClear:
	default:
		return nil, fmt.Errorf("dmx: unknown deMode %q (want %q, %q or %q)", c.DEMode, deModeNone, deModeAssert, deModeClear)
	}

	profiles := make(map[string]Profile, len(builtinProfiles)+len(c.Profiles))
	for k, v := range builtinProfiles {
		profiles[k] = v
	}
	for k, v := range c.Profiles {
		profiles[k] = v
	}

	resolvedProfiles := make(map[string]resolvedProfile, len(profiles))
	for name, p := range profiles {
		rp, err := resolveProfile(name, p)
		if err != nil {
			return nil, err
		}
		resolvedProfiles[name] = rp
	}

	fixtureKeys := sortedKeys(c.Fixtures)
	fixtures := make(map[string]resolvedFixture, len(c.Fixtures))
	for _, key := range fixtureKeys {
		fx := c.Fixtures[key]
		rp, ok := resolvedProfiles[fx.Profile]
		if !ok {
			return nil, fmt.Errorf("dmx: fixture %q: unknown profile %q", key, fx.Profile)
		}
		if fx.Address < 1 {
			return nil, fmt.Errorf("dmx: fixture %q: address must be >= 1", key)
		}
		end := fx.Address - 1 + len(rp.channels) - 1
		if end >= dmxUniverseChannels {
			return nil, fmt.Errorf("dmx: fixture %q: address %d with profile %q overflows the 512-channel universe", key, fx.Address, fx.Profile)
		}
		fixtures[key] = resolvedFixture{address: fx.Address, profile: rp}
	}

	if err := checkOverlaps(fixtureKeys, fixtures); err != nil {
		return nil, err
	}

	universeSize := minUniverseSize
	for _, fx := range fixtures {
		_, end := fx.span()
		if end+1 > universeSize {
			universeSize = end + 1
		}
	}
	if c.UniverseSize > 0 {
		// A decoder addressed past the frame this node sends never receives
		// its channels at all, so an explicit universeSize (e.g. to always
		// send the full 512-channel universe, like a real controller does)
		// is allowed as long as it still covers every declared fixture.
		if c.UniverseSize > dmxUniverseChannels {
			return nil, fmt.Errorf("dmx: universeSize must be <= %d", dmxUniverseChannels)
		}
		if c.UniverseSize < universeSize {
			return nil, fmt.Errorf("dmx: universeSize %d is smaller than the %d channels used by configured fixtures", c.UniverseSize, universeSize)
		}
		universeSize = c.UniverseSize
	}
	if universeSize > dmxUniverseChannels {
		universeSize = dmxUniverseChannels
	}

	// A full-universe frame takes real transmission time on this
	// software-timed cable: (universeSize+1 bytes)*44us, plus break/MAB
	// overhead. If fps demands a shorter interval than that, renderFrame's
	// own Send() call will run long and the frame ticker (buffered 1) drops
	// ticks, silently degrading the real output rate. Clamp fps down instead
	// of letting that happen invisibly.
	frameBytes := time.Duration(universeSize+1) * microsPerDataByte * time.Microsecond
	minFrameInterval := frameBytes + minBreakOverhead
	if maxAchievableFPS := int(time.Second / minFrameInterval); maxAchievableFPS < fps {
		if maxAchievableFPS < 1 {
			maxAchievableFPS = 1
		}
		logrus.Warnf("dmx: fps %d is unachievable with a %d-channel universe (min frame time %s); clamping to %d", fps, universeSize, minFrameInterval, maxAchievableFPS)
		fps = maxAchievableFPS
	}

	groups := make(map[string]*resolvedGroup, len(c.Groups))
	for _, key := range sortedKeys(c.Groups) {
		rg, err := resolveGroup(key, c.Groups[key], fixtures)
		if err != nil {
			return nil, err
		}
		groups[key] = rg
	}

	return &resolvedConfig{
		port:         c.Port,
		fps:          fps,
		breakMode:    breakMode,
		deMode:       deMode,
		fixtures:     fixtures,
		groups:       groups,
		universeSize: universeSize,
	}, nil
}

func resolveProfile(name string, p Profile) (resolvedProfile, error) {
	if len(p.Channels) == 0 {
		return resolvedProfile{}, fmt.Errorf("dmx: profile %q: must declare at least one channel", name)
	}

	channelIndex := make(map[string]int, len(p.Channels))
	for i, c := range p.Channels {
		if c == "" {
			return resolvedProfile{}, fmt.Errorf("dmx: profile %q: channel %d has empty name", name, i)
		}
		channelIndex[c] = i
	}

	static := make(map[int]byte, len(p.Static))
	for role, v := range p.Static {
		offset, ok := channelIndex[role]
		if !ok {
			return resolvedProfile{}, fmt.Errorf("dmx: profile %q: static channel %q is not declared in channels", name, role)
		}
		if v < 0 || v > 255 {
			return resolvedProfile{}, fmt.Errorf("dmx: profile %q: static value for %q must be 0-255", name, role)
		}
		static[offset] = byte(v)
	}

	return resolvedProfile{channels: p.Channels, static: static}, nil
}

// checkOverlaps ensures no two fixtures share a DMX channel. keys must be
// sorted so error messages are deterministic.
func checkOverlaps(keys []string, fixtures map[string]resolvedFixture) error {
	for i, a := range keys {
		as, ae := fixtures[a].span()
		for _, b := range keys[i+1:] {
			bs, be := fixtures[b].span()
			if as <= be && bs <= ae {
				return fmt.Errorf("dmx: fixtures %q and %q overlap on DMX channels", a, b)
			}
		}
	}
	return nil
}

func resolveGroup(key string, g Group, fixtures map[string]resolvedFixture) (*resolvedGroup, error) {
	if len(g.Fixtures) == 0 {
		return nil, fmt.Errorf("dmx: group %q: must reference at least one fixture", key)
	}
	for _, fk := range g.Fixtures {
		if _, ok := fixtures[fk]; !ok {
			return nil, fmt.Errorf("dmx: group %q: unknown fixture %q", key, fk)
		}
	}

	pattern := g.Pattern
	if pattern == "" {
		pattern = "static"
	}
	if _, ok := patterns[pattern]; !ok {
		return nil, fmt.Errorf("dmx: group %q: unknown pattern %q", key, pattern)
	}

	interval := time.Duration(g.Interval)
	if interval == 0 {
		interval = defaultInterval
	}
	if interval < 0 {
		return nil, fmt.Errorf("dmx: group %q: interval must be positive", key)
	}

	colorStrs := g.Colors
	if len(colorStrs) == 0 {
		colorStrs = []string{"#ffffff"}
	}
	colors := make([]rgb, 0, len(colorStrs))
	for _, cs := range colorStrs {
		c, err := parseHexColor(cs)
		if err != nil {
			return nil, fmt.Errorf("dmx: group %q: %w", key, err)
		}
		colors = append(colors, c)
	}

	name := g.Name
	if name == "" {
		name = key
	}

	return &resolvedGroup{
		name:     name,
		fixtures: g.Fixtures,
		pattern:  pattern,
		interval: interval,
		colors:   colors,
		reverse:  g.Reverse,
	}, nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
