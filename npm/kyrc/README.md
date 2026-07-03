# kyrc

A fast, offline, keyboard-only typing test that lives in your terminal.

Monkeytype is a website. **kyrc** is already where you are — it starts in
milliseconds, works with no network, and never asks you to log in.

## Install

```sh
npm i -g @kyrc/kyrc
# or: bun add -g @kyrc/kyrc   ·   pnpm add -g @kyrc/kyrc
```

The package is scoped `@kyrc/kyrc`, but the installed command is just `kyrc`.
A prebuilt static binary for your platform is delivered automatically (macOS
Intel/Apple Silicon, Linux x64/arm64, Windows x64) — no Go toolchain needed.

## Usage

```sh
kyrc            # 25-word test (default)
kyrc -w 50      # 50-word test
kyrc -t 30      # 30-second test
kyrc -t 1m      # 1-minute test
kyrc -q         # random quote
```

Keys: type to start · `backspace` delete · `ctrl+w` delete word ·
`tab` restart · `esc` / `ctrl+c` quit.

## What it measures

| stat | definition |
| --- | --- |
| **wpm** | correct characters ÷ 5 ÷ minutes (5-char-word convention, Monkeytype-style) |
| **raw** | all typed characters ÷ 5 ÷ minutes, ignoring correctness |
| **acc** | correct keystrokes ÷ total keystrokes (a mistyped-then-fixed char still counts as an error) |
| **consistency** | `1 − CV` of per-second raw WPM — higher is steadier |

The clock starts on your **first keystroke**, so idle time never counts, and
pasting is rejected so it can't inflate WPM.

## How it's built

A static Go binary with a [Bubble Tea](https://github.com/charmbracelet/bubbletea)
UI over a pure, headless typing engine — metrics are computed as pure functions
of a timestamped keystroke log, so every run is reproducible and auditable.

Source, issues, and full docs:
**[github.com/abh1nav9/kyrc](https://github.com/abh1nav9/kyrc)**

## License

MIT
