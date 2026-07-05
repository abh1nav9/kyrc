# kyrc

A fast, offline, keyboard-only typing test that lives in your terminal.

Monkeytype is a website. `kyrc` is already where you are — it starts in
milliseconds, works with no network, and never asks you to log in.

```
kyrc            # 25-word test (default)
kyrc -w 50      # 50-word test
kyrc -t 30      # 30-second test
kyrc -t 1m      # 1-minute test
kyrc -q         # random quote

kyrc results    # your last 10 results (sort by top / low WPM)
kyrc login      # create or restore a leaderboard account
kyrc whoami     # show your user_id + where your passkey is saved
kyrc leaderboard# view the global leaderboard
kyrc sync       # push your best result now
```

Keys: type to start · `backspace` delete · `ctrl+w` delete word ·
`tab` restart · `esc`/`ctrl+c` quit.

## Install

The canonical package is **`@kyrc/kyrc`** on npm. Every channel installs the
same static binary, and the installed command is always `kyrc`.

```sh
# Any platform with Node — npm / bun / pnpm (recommended)
npm i -g @kyrc/kyrc
bun add -g @kyrc/kyrc
pnpm add -g @kyrc/kyrc
npx @kyrc/kyrc            # run without installing

# macOS / Linux — Homebrew
brew install abh1nav9/tap/kyrc

# Windows — Scoop
scoop bucket add kyrc https://github.com/abh1nav9/scoop-bucket
scoop install kyrc

# Windows — WinGet
winget install abh1nav9.kyrc

# Debian/Ubuntu (apt) and Fedora/RHEL (dnf)
#   → https://abh1nav9.github.io/kyrc/

# Or grab a static binary, .deb, or .rpm from the releases page.
```

### In progress

These channels are wired up but not publishing yet:

- **AUR** (`yay -S kyrc-bin`) — Arch is holding new AUR registrations during
  the "Atomic Arch" supply-chain cleanup; enabled the moment it reopens.
- **Snap** (`snap install kyrc`) — Snap Store account/name registration
  pending.

The npm package is scoped `@kyrc/kyrc` (the unscoped name `kyrc` is blocked by
npm's name-similarity filter), but the installed command is still just
`kyrc`. For how the release pipeline publishes to every channel from one git
tag — and the one-time setup each needs — see
**[docs/DISTRIBUTION.md](docs/DISTRIBUTION.md)**.

## Results, accounts & the leaderboard

kyrc keeps working with **no account and no network** — that never changes.
Every test you take is saved locally (your last 10), and you can review them
sorted by your best or worst runs:

```sh
kyrc results     # opens a sortable view: [t] top wpm · [l] low wpm · [r] recent
```

When you *want* to compete, opt into an account. kyrc gives you a **user_id**
and a **passkey** (a recovery phrase). Your private key is generated on your
machine and **never leaves it**; scores are signed locally and the server
**replays your keystroke log** to verify the WPM — so nobody can fake a score
or submit as you.

```sh
kyrc login "Your Name"    # create an account → prints your user_id + passkey
kyrc whoami               # show your user_id + the path to your recovery file
kyrc leaderboard          # view the global top typists
kyrc sync                 # push your best now (also auto-syncs after tests)
```

### Where to find your user_id and passkey

`kyrc whoami` prints them, and a private recovery card is saved at:

| OS      | Path |
|---------|------|
| macOS   | `~/Library/Application Support/kyrc/recovery.txt` |
| Linux   | `~/.config/kyrc/recovery.txt` |
| Windows | `%AppData%\kyrc\recovery.txt` |

> Keep this file private — anyone with the passkey can log in as you. kyrc
> never transmits it; it's only used locally to re-derive your key.

### Logging in again (new machine)

```sh
kyrc login          # choose "restore", then enter your user_id + passkey
```

This rebuilds the **same** account — same user_id, same leaderboard identity.
Because the user_id is a fingerprint of your public key, restoring from the
passkey always reproduces it exactly.

### How the anti-cheat works (the short version)

- **Accounts can't be spoofed.** Auth is an Ed25519 signature, not a shared
  secret. The server stores only your *public* key and verifies signatures.
  Knowing someone's user_id (it's public) grants zero power to act as them.
- **Scores can't be faked.** Submissions carry the raw, timestamped keystroke
  log. The server recomputes WPM from it (deriving elapsed time from the log's
  own timestamps) and rejects anything that doesn't match.
- **Requests can't be replayed.** A per-submission nonce + timestamp window and
  a per-run digest stop captured requests and double-submits.

The leaderboard service lives in [`server/`](server/); it's the only component
that holds the database URL. See [server/README.md](server/README.md).

## Why it feels instant

> For the full engineering story — architecture, decisions, and the hurdles
> we hit (npm binary distribution, 2FA publishing, PTY testing, and more) —
> see **[docs/ENGINEERING.md](docs/ENGINEERING.md)**.

kyrc is engineered as a real-time input→feedback loop, not a form:

- **The clock is owned deliberately.** Every keystroke is timestamped at
  capture. The typing engine is a pure function of that timestamped event
  stream — no clock is ever read inside the logic — so a run is fully
  reproducible and replayable.
- **The engine knows nothing about terminals.** `internal/engine` is a
  headless finite state machine (`Idle → Running → Finished`) with metrics
  computed as pure functions over an append-only keystroke log. It's unit-
  tested with synthetic timestamps; the UI is a thin
  [Bubble Tea](https://github.com/charmbracelet/bubbletea) adapter over it.
- **Feedback and the clock refresh independently.** Per-keystroke feedback
  is immediate; the countdown redraws at ~15 Hz. A late clock frame is
  invisible; a late keystroke is not.

## Metrics (how the numbers are defined)

- **wpm** (hero number): correct characters ÷ 5 ÷ minutes — the standard
  5-characters-per-word convention, matching Monkeytype's headline.
- **raw**: all typed characters ÷ 5 ÷ minutes, ignoring correctness.
- **acc**: correct keystrokes ÷ total keystrokes. This is *keystroke*
  accuracy — a character you mistyped and fixed still counts as an error,
  even though the final text is clean.
- **consistency**: `1 − CV` of per-second raw WPM. Higher = steadier.
- The clock starts on the **first keystroke**, never on test render, so
  idle time never counts. Pasting is rejected so it can't inflate WPM.

Because metrics are pure functions over the keystroke log, any result can
be recomputed and audited from the log alone.

## Architecture

```
cmd/kyrc            CLI entry, flag parsing, subcommands, version metadata
internal/engine     pure state machine + metrics (no terminal, fully tested)
internal/wordsource word/quote generation behind a Source interface
internal/input      terminal key → engine event translation (+ timestamps)
internal/ui         Bubble Tea adapter: typing + results screens
internal/store      local results history (last 10), offline, atomic writes
internal/identity   Ed25519 keypair, user_id, recovery phrase, key files
internal/leaderboard signed submission protocol, replay anti-cheat, sync client
server/             leaderboard API in front of Neon/Postgres (its own module)
npm/                esbuild-style npm distribution (meta pkg + platform pkgs)
site/               React landing page (install, leaderboard, account docs)
.goreleaser.yaml    cross-platform static binaries + GitHub releases
```

## Building & releasing

```sh
make build       # local binary with version metadata
make test        # run the suite
make run         # build + run
make snapshot    # cross-platform binaries via GoReleaser (no publish)
make npm-stage   # stage npm platform packages from dist/
```

Releases are cut from a git tag: GoReleaser builds the static
(`CGO_ENABLED=0`) matrix for darwin/linux/windows × amd64/arm64 and publishes
a GitHub release with checksums. The npm platform packages are staged from
the same `dist/` and published to npm.

## License

MIT
