// Command kyrc is a terminal typing test. First run with no flags starts a
// test instantly — no config, no login. That instant first experience is
// the product.
package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/abh1nav9/kyrc/internal/engine"
	"github.com/abh1nav9/kyrc/internal/ui"
	"github.com/abh1nav9/kyrc/internal/wordsource"
)

// Build metadata, injected via -ldflags at release time so bug reports are
// actionable. Defaults identify a local/dev build.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Subcommands are opt-in extras. A bare `kyrc` (or any flags) still
	// starts a test instantly — the instant, no-login first run is the
	// product, so identity/leaderboard live behind explicit verbs.
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "results", "history":
			runResults()
			return
		case "login", "account":
			runLogin(os.Args[2:])
			return
		case "whoami":
			runWhoami()
			return
		case "leaderboard", "board":
			runLeaderboard()
			return
		case "sync":
			runSync()
			return
		}
	}

	cfg, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "kyrc:", err)
		os.Exit(2)
	}

	m := ui.New(cfg)
	m.OnFinish = saveAndSync // persist the result + best-effort leaderboard sync
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "kyrc:", err)
		os.Exit(1)
	}
}

func parseArgs(args []string) (ui.Config, error) {
	// Defaults: 25-word test on the random English source.
	cfg := ui.Config{
		Mode:      engine.ModeWords,
		WordCount: 25,
		Source:    wordsource.NewRandom(rand.New(rand.NewSource(time.Now().UnixNano()))),
	}

	i := 0
	for i < len(args) {
		a := args[i]
		switch {
		case a == "-h" || a == "--help":
			printUsage()
			os.Exit(0)
		case a == "-v" || a == "--version":
			fmt.Printf("kyrc %s (commit %s, built %s)\n", version, commit, date)
			os.Exit(0)
		case a == "-t" || a == "--time":
			v, rest, err := takeValue(args, i)
			if err != nil {
				return cfg, err
			}
			d, err := parseDuration(v)
			if err != nil {
				return cfg, err
			}
			cfg.Mode = engine.ModeTime
			cfg.Duration = d
			i = rest
		case a == "-w" || a == "--words":
			v, rest, err := takeValue(args, i)
			if err != nil {
				return cfg, err
			}
			n, err := parsePositiveInt(v)
			if err != nil {
				return cfg, err
			}
			cfg.Mode = engine.ModeWords
			cfg.WordCount = n
			i = rest
		case a == "-q" || a == "--quote":
			cfg.Source = wordsource.Static{Text: randomQuote()}
			cfg.Mode = engine.ModeWords
			cfg.WordCount = 0 // Static ignores count
			i++
		default:
			return cfg, fmt.Errorf("unknown flag %q (try --help)", a)
		}
	}
	return cfg, nil
}

func takeValue(args []string, i int) (string, int, error) {
	if i+1 >= len(args) {
		return "", i, fmt.Errorf("flag %q needs a value", args[i])
	}
	return args[i+1], i + 2, nil
}

func parseDuration(s string) (time.Duration, error) {
	// Bare number means seconds; otherwise Go duration syntax (e.g. 1m30s).
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}
	if n, err := parsePositiveInt(s); err == nil {
		return time.Duration(n) * time.Second, nil
	}
	return 0, fmt.Errorf("bad duration %q (try 30 or 1m)", s)
}

func parsePositiveInt(s string) (int, error) {
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("not a number: %q", s)
		}
		n = n*10 + int(r-'0')
	}
	if n <= 0 {
		return 0, fmt.Errorf("must be positive: %q", s)
	}
	return n, nil
}

func printUsage() {
	fmt.Print(strings.TrimLeft(`
kyrc — a fast terminal typing test

usage:
  kyrc                 25-word test (default)
  kyrc -w 50           50-word test
  kyrc -t 30           30-second test
  kyrc -t 1m           1-minute test
  kyrc -q              random quote

keys:
  type       start the test on first keystroke
  backspace  delete   ·   ctrl+w delete word
  tab        restart  ·   esc / ctrl+c quit

flags:
  -w, --words N     number of words
  -t, --time DUR    timed test (seconds, or 30s/1m)
  -q, --quote       quote mode
  -v, --version     print version
  -h, --help        this help
`, "\n"))
}

var quotes = []string{
	"The only way to do great work is to love what you do.",
	"Simplicity is the ultimate sophistication.",
	"Programs must be written for people to read, and only incidentally for machines to execute.",
	"Premature optimization is the root of all evil.",
	"Talk is cheap. Show me the code.",
}

func randomQuote() string {
	return quotes[rand.Intn(len(quotes))]
}
