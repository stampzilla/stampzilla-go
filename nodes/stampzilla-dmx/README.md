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
`termios2`/`BOTHER` (see `opendmx_linux.go`). The break itself can be
generated two different ways (see `breakMode` below) — which one actually
works reliably is adapter/kernel-driver dependent and isn't something that
can be predicted in advance.

Some RS485 cables also wire the transceiver's driver-enable (DE) pin to the
FT232's RTS or DTR line; whether that pin needs to be asserted, cleared, or
left alone to enable the line driver also varies by cable (see `deMode`
below).

### Standalone bring-up / troubleshooting

If a decoder isn't responding — especially if its address buttons lock up as
soon as this node starts, which usually means *something* is arriving on the
line, just not a valid or expected DMX signal — bypass the server entirely
and drive the cable directly:

```
stampzilla-dmx selftest -port /dev/ttyUSB0 -mode full
```

This is a **subcommand**, not a flag on the main command line — the node's
own flags (`-host`, `-port`, `-loglevel`, ...) are parsed by a separate
mechanism deep inside node startup that has no way to learn about extra
flags, so selftest options live behind the leading `selftest` argument
instead of colliding with them. Running `stampzilla-dmx` (no subcommand)
still shows the normal node flags and env vars, unaffected.

This sends a continuous 255-on-every-channel frame with no config, no
server connection and no device state involved. Other modes:

- `-mode walk` — lights one channel at a time (2s each by default), logging
  which DMX slot is active, so a decoder's physical outputs can be mapped to
  addresses without reading its (possibly locked) front panel.
- `-mode ramp` — a slow fade across all channels, to check dimming
  independently of on/off.

Combine with `-break-mode ioctl|baud` and `-de-mode none|assert|clear` to
try every combination without rebuilding — see `breakMode`/`deMode` below,
which are the same settings the normal `port` config exposes. `-echo`
additionally reads from the port while transmitting and logs anything seen,
which on some cables confirms the framing made it onto the wire
(best-effort: most transceivers won't echo their own output). Run
`stampzilla-dmx selftest -h` for the full flag list.

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
- `fps` — frame rate the universe is refreshed at, defaults to 30, clamped to
  44, and further clamped down if it's unachievable for the configured
  `universeSize` (a full 512-channel frame takes real transmission time on
  this software-timed cable).
- `universeSize` — number of channels sent every frame, 24-512. Defaults to
  just enough to cover the highest-addressed configured fixture (floored at
  24). Set explicitly (e.g. to 512) to always send the full universe, like a
  real DMX controller does — otherwise a fixture addressed above the
  smallest frame that happens to cover your other fixtures will never
  receive its channels.
- `breakMode` — how the BREAK/MAB is generated: `baud` (default; drops to
  50,000 baud and writes a single `0x00` byte) or `ioctl` (`TIOCSBRK`/
  `TIOCCBRK` with fixed sleeps). Both are legitimate for a bare FTDI/RS485
  cable; which one a given adapter/kernel combination handles cleanly isn't
  predictable, so it's a config knob rather than fixed — see "Standalone
  bring-up" above to try both against real hardware.
- `deMode` — whether to touch the RTS/DTR modem control lines when opening
  the port: `none` (default; never touch them), `assert`, or `clear`. Only
  relevant for RS485 cables whose transceiver DE pin is wired to RTS or DTR;
  a failure here (e.g. an adapter that doesn't support it) is logged and
  otherwise ignored, never fatal.
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
    (`chase`, `scanner`, `fill`, `fillonce`, `alternate`, `wave`, `rainbow`).

## Patterns

| Pattern | Behavior | Colors used |
| --- | --- | --- |
| `off` | Every fixture stays dark. | — |
| `static` | Every fixture at full brightness. | First color only |
| `chase` | One fixture lit at a time, moving through the group in order. | Cycles through colors, one per step |
| `scanner` | One fixture lit at a time, sweeping back and forth across the group (Knight Rider/KITT-style) instead of wrapping like `chase` — each end is visited once per bounce, no pause or double-hit. | First color only |
| `fill` | Fixtures light up one at a time (0, 1, 2, ...) until the whole group is on, then turn off one at a time starting from the last one lit, then repeat. | First color only |
| `fillonce` | Like `fill`, but doesn't loop: turning the group **on** fills fixtures in order once and holds them lit; turning it **off** plays the same fill in reverse (last fixture first) once and holds everything dark. Each direction takes `channel count × interval`. See note below — this is the only pattern where turning a group off is animated instead of instant. | First color only |
| `alternate` | Every other fixture lit; which half is lit flips each step. | Cycles through colors, one per step |
| `pulse` | All fixtures breathe (fade in and out) together. | First color only |
| `wave` | Like `pulse`, but each fixture is phase-offset by its position, producing a wave that travels across the group. | First color only |
| `colorcycle` | All fixtures at full brightness, stepping through the configured colors together. | Cycles through colors, one per step |
| `rainbow` | Hue sweeps across the full color wheel, offset by each fixture's position in the group. | Ignored — uses an HSV hue sweep instead |
| `random` | Deterministic pseudo-random intensity and color per fixture, changing every step (reproducible, not `math/rand`). | Picked pseudo-randomly per fixture/step |

`reverse` flips fixture order for the positional patterns (`chase`, `scanner`,
`fill`, `fillonce`, `alternate`, `wave`, `rainbow`) — the others ignore
fixture position and are unaffected. For `fillonce`, `reverse` also flips
which end each direction starts from; for `scanner`, it just flips which end
the sweep starts from.

Fixtures without a `dimmer` channel have their color channels scaled directly
by intensity/brightness; fixtures with a `dimmer` channel keep color channels
at full value and let `dimmer` carry the level. `white` (on `rgbw`-style
profiles) is derived from the minimum of the pattern's red/green/blue.

Every pattern except `fillonce` stops being rendered the instant a group is
switched off — channels simply go to (or stay at) whatever a fixture's
profile `static` values say, since the group is skipped entirely. `fillonce`
is the one exception: it keeps rendering (at intensity 0, counting down)
while it plays its off animation, so a `static` value on a color/dimmer role
channel is temporarily overwritten to 0 during and after that drain, while
`static` values on other roles (e.g. `mode`, `amber`, `uv`) are unaffected.
Interrupting a `fillonce` animation partway through and reversing it (e.g.
toggling off after only a few fixtures have filled in) resumes the new
direction from the currently-visible fixture count rather than restarting
from the opposite extreme, so it never flashes extra fixtures first.
