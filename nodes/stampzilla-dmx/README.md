# dmx

Drives DMX-512 fixtures (LED pars, dimmer packs, etc.) through an **Open DMX
USB**-style cable connected as a serial port. Fixtures are addressed
individually in config, but only **groups** of fixtures are exposed as
stampzilla devices — a group runs one of several built-in animated patterns
(chase, pulse, rainbow, ...) across its fixtures.

Each group becomes a `light` device with `OnOff` and `Brightness` traits,
plus a plain `pattern` state key that can be changed at runtime (from a
saved-state/scene, the device's JSON editor in the dashboard, or the
websocket API) to switch which pattern the group is running.

If `port` is left empty, the node logs the frames it would have sent instead
of writing to hardware — useful for testing a config without a cable
attached.

## Hardware

This node targets **Open DMX USB**-style cables: a bare FTDI USB-to-serial
chip plus an RS485 line driver, with no onboard microcontroller (these are
sold generically as "USB to DMX RS485 FTDI cable" and similar). The node
generates the DMX-512 line timing itself (break, mark-after-break, 250,000
baud 8N2) directly over the serial port — Linux only, using raw
`termios2`/`BOTHER` and break ioctls (see `opendmx_linux.go`).

**It does not speak the Enttec DMX USB Pro / DMXKing UltraDMX Widget API
protocol**, so it will not work with those (or similarly "smart",
firmware-based) widgets — only with the dumb, wire-only cables described
above.

Because the DMX timing is generated in software rather than by dedicated
widget firmware, it's less rock-solid than a "smart" widget: `time.Sleep` at
the ~100µs scale has real jitter on a general-purpose Linux scheduler, and
the FTDI kernel driver's default 16ms USB latency timer can add per-write
lag. Most DMX receivers tolerate this fine, but if you see flicker or
inconsistent timing, try reducing the latency timer for the port:

```
echo 1 | sudo tee /sys/bus/usb-serial/devices/ttyUSB0/latency_timer
```

(add a udev rule to make this persist across reconnects).

## Configuration

```json
{
  "port": "/dev/ttyUSB0",
  "fps": 30,
  "profiles": {
    "mypar": {
      "channels": ["mode", "dimmer", "red", "green", "blue", "white"],
      "static": { "mode": 255 }
    }
  },
  "fixtures": {
    "par1": { "profile": "mypar", "address": 1 },
    "par2": { "profile": "mypar", "address": 7 },
    "par3": { "profile": "mypar", "address": 13 },
    "spot": { "profile": "dimmer", "address": 100 }
  },
  "groups": {
    "stage": {
      "name": "Stage wash",
      "fixtures": ["par1", "par2", "par3"],
      "pattern": "chase",
      "interval": "400ms",
      "colors": ["#ff0000", "#00ff00", "#0000ff"],
      "reverse": false
    },
    "spots": {
      "name": "Spots",
      "fixtures": ["spot"],
      "pattern": "pulse",
      "interval": "2s"
    }
  }
}
```

- `port` — serial device for the DMX cable. Empty logs frames instead of
  sending them. Always driven at a fixed 250,000 baud, 8N2 — not configurable.
- `fps` — frame rate the universe is refreshed at, defaults to 30, clamped to 44.
- `profiles` — named channel layouts. Each entry in `channels` is a role,
  applied at increasing offsets from a fixture's `address`. Recognised roles:
  `dimmer`, `red`, `green`, `blue`, `white`, `amber`, `uv`. Any other name
  (including `-`) is an opaque channel only ever driven by `static`. Built-in
  profiles (usable without declaring them): `dimmer`, `rgb`, `rgbw`, `rgba`,
  `rgb-dimmer`, `rgbw-dimmer`. A profile declared under the same name as a
  built-in overrides it.
- `fixtures` — places a `profile` at a DMX start `address` (1-512).
- `groups` — a set of `fixtures` (by key) driven together:
  - `pattern` — see [Patterns](#patterns) below. Defaults to `static`.
  - `interval` — how often the pattern advances one step, e.g. `"400ms"`,
    `"2s"`. Defaults to `500ms`.
  - `colors` — hex colors (`#rgb` or `#rrggbb`) the pattern draws from.
    Defaults to `["#ffffff"]`.
  - `reverse` — flips fixture order for the positional patterns
    (`chase`, `fill`, `alternate`, `wave`, `rainbow`).

## Patterns

| Pattern | Behavior | Colors used |
| --- | --- | --- |
| `off` | Every fixture stays dark. | — |
| `static` | Every fixture at full brightness. | First color only |
| `chase` | One fixture lit at a time, moving through the group in order. | Cycles through colors, one per step |
| `fill` | Fixtures light up one at a time (0, 1, 2, ...) until the whole group is on, then turn off one at a time starting from the last one lit, then repeat. | First color only |
| `alternate` | Every other fixture lit; which half is lit flips each step. | Cycles through colors, one per step |
| `pulse` | All fixtures breathe (fade in and out) together. | First color only |
| `wave` | Like `pulse`, but each fixture is phase-offset by its position, producing a wave that travels across the group. | First color only |
| `colorcycle` | All fixtures at full brightness, stepping through the configured colors together. | Cycles through colors, one per step |
| `rainbow` | Hue sweeps across the full color wheel, offset by each fixture's position in the group. | Ignored — uses an HSV hue sweep instead |
| `random` | Deterministic pseudo-random intensity and color per fixture, changing every step (reproducible, not `math/rand`). | Picked pseudo-randomly per fixture/step |

`reverse` flips fixture order for the positional patterns (`chase`, `fill`,
`alternate`, `wave`, `rainbow`) — the others ignore fixture position and are
unaffected.

Fixtures without a `dimmer` channel have their color channels scaled directly
by intensity/brightness; fixtures with a `dimmer` channel keep color channels
at full value and let `dimmer` carry the level. `white` (on `rgbw`-style
profiles) is derived from the minimum of the pattern's red/green/blue.
