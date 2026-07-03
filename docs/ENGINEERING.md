# kyrc — Engineering Notes

How kyrc is built, the decisions behind it, and the problems we hit getting
it onto npm. This is the "why," not the "what" — for usage see the
[README](../README.md).

---

## 1. The core principle

kyrc is not a form that collects text and prints a score. It is a **real-time
human–computer feedback loop**, and everything in the architecture follows
from one sentence we kept on a sticky note:

> **Own the clock, keep the engine pure, and make input→pixel feel free.**

Three consequences:

- **Latency perception beats throughput.** A tool can process 10,000
  keystrokes/sec and still *feel* laggy if each one takes 120 ms to paint.
  What matters is the time from a keypress to the pixel that acknowledges it.
- **Trust beats cleverness.** Users compare their WPM to Monkeytype and lose
  faith instantly if it's off. Every number must be defined precisely and be
  reproducible.
- **Determinism beats convenience.** If the same input always produces the
  same output, the whole system becomes testable, debuggable, and auditable.

---

## 2. Architecture at a glance

The codebase is split by a single hard rule: **the engine knows nothing about
terminals, and the UI knows nothing about scoring math.** This is
ports-and-adapters (hexagonal) architecture — a pure core with the terminal
as one replaceable adapter.

```
cmd/kyrc            CLI entry, flag parsing, version metadata (via ldflags)
internal/engine     pure state machine + metrics — NO terminal, fully tested
internal/wordsource word / quote generation behind a Source interface
internal/input      terminal key → engine event translation (+ timestamps)
internal/ui         Bubble Tea adapter: renders the engine, owns no state math
npm/                esbuild-style npm distribution (meta pkg + platform pkgs)
site/               React + Tailwind + Motion landing page
.github/workflows   tag-driven release automation
```

Data flows one way:

```
keypress ─▶ internal/input ─▶ engine.Event{Kind, Rune, At} ─▶ engine.Session.Apply
                (timestamp                                          │
                 stamped here)                                      ▼
                                                        append-only []Keystroke
                                                                    │
   render ◀── internal/ui.View(session) ◀── engine state ◀─────────┘
                (pure function of state)        engine.Compute(log) ─▶ Metrics
```

The engine is ~1,000 lines with **14 unit tests** that run with zero terminal
involvement (and under Go's race detector). If we can't compute a correct WPM
from a fake, timestamped event log, nothing else matters — so that came first.

---

## 3. Key engineering decisions

### 3.1 Time is injected, never read

The single most important rule in the engine: **it never calls
`time.Now()`.** Instead, every input carries the moment it was captured:

```
type Event struct {
    Kind KeyKind
    Rune rune
    At   time.Time   // the ONLY clock the engine sees
}
```

**Why this matters:** if the engine sampled the clock internally, its behavior
would depend on *when it happened to run* — scheduler jitter, GC pauses, the
Bubble Tea event loop's batching would all leak into your WPM. By taking time
as data that arrives with the event, the engine becomes a pure function of an
ordered event stream. Replaying the same events always yields identical state
and identical metrics. Tests feed synthetic timestamps and assert exact
numbers (e.g. "25 chars in 6 s = exactly 50 WPM").

The timestamp is captured in `internal/input`, as close to the keypress as we
control, so metrics inherit as little jitter as possible.

### 3.2 The session is an explicit finite state machine

A test is modeled as `Idle → Running → Finished` (timed tests can also finish
on a deadline). `Apply(event)` is the *only* mutator. Drawing this as an
explicit FSM — rather than scattering booleans like `hasStarted`,
`isDone` — killed an entire class of "what happens if the user backspaces
before typing anything / types after the timer expired / pastes on the results
screen" bugs. Each is now a defined transition (or an explicit no-op).

**The clock starts on the first keystroke, never on test render.** Idle time
before the user starts typing must not count — this is the near-universal
convention and the difference between a trusted and a "broken" WPM.

### 3.3 Metrics are pure functions over an append-only log

Rather than mutating live counters (`correctCount++`), every keystroke is
appended to a `[]Keystroke` log, and WPM / accuracy / consistency are computed
by folding over that log on demand:

```
Compute(log []Keystroke, elapsed) → Metrics
```

**Why:** this is a lightweight form of event sourcing. It makes every metric
reproducible and *auditable* — when someone says "your WPM is wrong," we can
replay the exact session and show the math. It also let us match Monkeytype's
definitions deliberately rather than by accident:

| stat | definition |
| --- | --- |
| **wpm** (hero) | correct chars ÷ 5 ÷ minutes (the 5-char-word convention) |
| **raw** | all typed chars ÷ 5 ÷ minutes, ignoring correctness |
| **acc** | correct keystrokes ÷ total keystrokes — *keystroke* accuracy, so a char you mistyped and fixed still counts as an error |
| **consistency** | `1 − CV` of per-second raw WPM (coefficient of variation) — rewards a steady pace |

The subtle one is **accuracy**: it measures *keystroke* correctness, not
final-text correctness. Typing `x`, backspacing, then typing the right letter
is 100% correct final text but counts one error. That's intentional and
matches how the reference tools behave.

### 3.4 Two render clocks, deliberately separated

Per-keystroke feedback and the on-screen timer refresh on **different
schedules**:

- **Typing feedback is immediate** — the moment a key lands, the character's
  color updates. This is the latency the user actually feels; the budget is
  one 60 fps frame (~16 ms).
- **The countdown/elapsed clock ticks at ~15 Hz** (a 66 ms Bubble Tea tick).

**Why split them:** a clock frame that's one tick late is invisible; a
keystroke that's one frame late is not. Coupling them would either make
feedback wait for the tick (laggy) or force a full repaint on every keystroke
*and* every tick (wasteful, and a source of GC churn). Keeping them
independent gets instant feedback and a cheap clock.

### 3.5 The UI is a thin adapter, not the app

`View(session)` is a **pure function of engine state** — it computes a string
and never mutates anything. All the WPM math lives in the engine; the UI just
renders whatever the engine says. This boundary is what lets the engine be
unit-tested headlessly, and it means the same core could back a web front-end
or a benchmark harness later without touching the scoring logic.

### 3.6 Static, zero-dependency binary

Built with `CGO_ENABLED=0`, so kyrc is a single static binary with no libc
dependency. **One artifact works on both glibc and musl (Alpine)** — that
alone eliminates a whole category of "works on my machine" support tickets.
This is Go's superpower for a CLI and the foundation of the distribution
story.

---

## 4. Problems we hit (and how we solved them)

### 4.1 Wrapping styled, variable-width text

**Problem:** the passage must wrap at the terminal width, but you can't wrap by
counting bytes — ANSI color codes and wide runes (CJK, emoji) each throw the
math off. Naive wrapping mis-positions the caret.

**Fix:** compute line breaks on the **plain** target text using *display
width* (`go-runewidth`), producing a set of "break before this index"
positions. Styling is applied afterward, so wrapping is correct regardless of
color codes or double-width characters.

### 4.2 Paste inflating WPM

**Problem:** a user (or an over-eager terminal) can dump a whole line into
stdin at once, which would register as superhumanly fast typing.

**Fix:** a single key message carrying **more than one rune** is treated as a
paste (or IME commit) and rejected — the engine never sees it, and the UI
flashes a brief "paste ignored" banner. This mirrors Monkeytype's behavior.

### 4.3 Completion firing before a mid-word correction

**Problem:** an early test used a 2-character target to check "mistype → fix →
retype." It failed because typing the 2nd character (even wrong) put the cursor
at the end, so the session finished *before* the backspace-and-fix could
happen.

**Fix / lesson:** this was a genuine correctness insight surfaced by a test —
in word mode, completion triggers the instant the cursor reaches the end of
the target. The test was rewritten around a longer target so the correction
happens mid-word, and the behavior itself is correct.

### 4.4 Verifying an interactive TUI without a human

**Problem:** the app needs raw-mode TTY input, so it can't be driven through a
normal pipe. How do you prove it renders and responds in CI-like conditions?

**Fix:** two layers. First, the Bubble Tea model is driven **headlessly** in
tests by feeding `KeyMsg` values straight through `Update` and asserting on
state — no terminal needed. Second, for a true end-to-end check we drove the
real binary inside a **pseudo-terminal (PTY)**. That surfaced a subtle gotcha:
the binary emits terminal *queries* (background color `\e]11;?`, cursor
position `\e[6n`) and **blocks waiting for the reply** before rendering the
first frame — a real terminal answers instantly, but our PTY harness had to
answer them explicitly. Once it did, we captured full styled frames and
confirmed the timed test reaches the results screen with every stat.

### 4.5 npm doesn't ship binaries — but kyrc *is* a binary

**Problem:** npm distributes JavaScript, but kyrc is a compiled Go binary.
Users expect `npm i -g kyrc` to just work on any OS/arch.

**Fix (the esbuild model):** a thin **meta-package** (`@kyrc/kyrc`, pure JS)
whose `optionalDependencies` are **per-platform packages**
(`@kyrc/darwin-arm64`, `@kyrc/linux-x64`, …). Each platform package declares
its own `os`/`cpu`, so npm installs **only the one** matching the host and
silently skips the other four. A tiny launcher (`bin/kyrc.js`) resolves
whichever landed and `exec`s the native binary with inherited stdio, so the
binary owns the TTY directly (essential for raw mode). Result:
`npm i -g @kyrc/kyrc` downloads one ~3 MB native binary for your platform and
nothing else.

### 4.6 The name `kyrc` is blocked on npm

**Problem:** publishing the unscoped `kyrc` failed with
*"Package name too similar to existing packages ky, rc, crc, nyc."* npm's
spam filter rejects short names close to existing ones.

**Fix:** scope the package as **`@kyrc/kyrc`** (we already owned the `@kyrc`
org). Crucially, the `bin` entry stays `kyrc`, so the **installed command is
still just `kyrc`** — only the install name changed.

### 4.7 2FA vs. non-interactive publish

**Problem:** the npm account enforces 2FA on writes. A plain CLI session (even
after `npm login`) can't satisfy that and returns *"Two-factor authentication
or granular access token with bypass 2fa enabled is required."*

**Fix:** publish with an **Automation token** (a classic token type that
bypasses 2FA by design, built for CI). We also learned to distrust npm's
optimistic `+ package@version` output — a publish can print success and still
be mid-propagation; the registry read-CDN lags the write by minutes, so
`curl`-ing the registry and polling is how you confirm a package is truly
live.

### 4.8 Keeping six package versions in lockstep

**Problem:** the meta-package pins each platform package by exact version. If
they drift, an install resolves a mismatched binary — or fails outright.
Bumping a version by hand across six `package.json` files is error-prone.

**Fix:** the **git tag is the single source of truth.** A release script reads
one `KYRC_VERSION`, stamps the meta-package, re-pins all its
`optionalDependencies`, and versions every platform package to match — so all
six always ship together. (A README-only change is the one exception: the
meta-package can bump alone while the unchanged binaries stay pinned to their
prior version.)

---

## 5. Distribution & release automation

Releases are **tag-driven**. A `git tag vX.Y.Z && git push --tags` triggers a
GitHub Actions workflow that:

1. Runs **GoReleaser** to cross-compile the static-binary matrix
   (darwin/linux/windows × amd64/arm64) and cut a GitHub Release with
   checksums. The tag drives the version via `-ldflags`, so
   `kyrc --version` matches the published npm version.
2. Stages the npm packages from that same `dist/` output and syncs all six
   versions to the tag.
3. Publishes the **platform packages first**, then the **meta-package** — so
   the meta-package's pinned dependencies already exist when it goes up —
   authenticating with an `NPM_TOKEN` repo secret (no local terminal, no OTP).

This turns a release from a multi-step, token-juggling manual process into one
command.

---

## 6. What we'd do next

- **Instrument real input→pixel latency** (p50/p99) rather than reasoning about
  the budget — "feeling fast" is a distribution, and p99 is what users
  remember. Test under tmux and over SSH, where the latency floor is real.
- **Persist results** to a local history file (still offline-first) for
  progress-over-time views.
- **More modes** — punctuation/numbers, custom word lists, quotes of varying
  length — all behind the existing `wordsource.Source` interface, so the
  engine never changes.
- **macOS notarization** so the raw binary download isn't Gatekeeper-blocked.
