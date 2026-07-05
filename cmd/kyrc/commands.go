package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/abh1nav9/kyrc/internal/engine"
	"github.com/abh1nav9/kyrc/internal/identity"
	"github.com/abh1nav9/kyrc/internal/leaderboard"
	"github.com/abh1nav9/kyrc/internal/store"
	"github.com/abh1nav9/kyrc/internal/ui"
)

// baseURL returns the leaderboard API base, overridable for self-hosting.
func baseURL() string {
	if v := os.Getenv("KYRC_LEADERBOARD_URL"); v != "" {
		return v
	}
	return leaderboard.DefaultBaseURL
}

// saveAndSync is the post-test hook: persist the result locally (always,
// offline) and, if the user has an identity, best-effort push their best to
// the leaderboard. Any network/identity error is swallowed — kyrc must keep
// working offline no matter what.
func saveAndSync(s *engine.Session, cfg ui.Config) {
	path, err := store.DefaultPath()
	if err != nil {
		return
	}
	h, err := store.Load(path)
	if err != nil {
		return
	}

	m := engine.Compute(s.Log(), s.Elapsed(time.Now()))
	mode, param := modeParam(cfg)
	r := store.FromMetrics(mode, param, m, s.Log(), time.Now())
	if err := h.Add(r); err != nil {
		return
	}

	// Best-effort leaderboard sync (only if logged in).
	dir, err := store.ConfigDir()
	if err != nil || !identity.Exists(dir) {
		return
	}
	id, err := identity.Load(dir)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	_ = leaderboard.NewClient(baseURL()).SyncBest(ctx, id, h) // ignore errors
}

func modeParam(cfg ui.Config) (string, int) {
	if cfg.Mode == engine.ModeTime {
		return "time", int(cfg.Duration.Seconds())
	}
	return "words", cfg.WordCount
}

// runResults renders the local history in a sortable Bubble Tea screen.
func runResults() {
	path, err := store.DefaultPath()
	if err != nil {
		fmt.Fprintln(os.Stderr, "kyrc:", err)
		os.Exit(1)
	}
	h, err := store.Load(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "kyrc:", err)
		os.Exit(1)
	}
	if err := ui.RunResults(h); err != nil {
		fmt.Fprintln(os.Stderr, "kyrc:", err)
		os.Exit(1)
	}
}

// runWhoami prints the current identity and where the recovery file lives.
func runWhoami() {
	dir, err := store.ConfigDir()
	if err != nil || !identity.Exists(dir) {
		fmt.Println("Not logged in. Run `kyrc login` to create an account.")
		return
	}
	name, userID, _, err := identity.LoadPublic(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "kyrc:", err)
		os.Exit(1)
	}
	_, _, recovery := identity.Paths(dir)
	fmt.Printf("name:     %s\n", name)
	fmt.Printf("user_id:  %s\n", userID)
	fmt.Printf("\nYour recovery phrase (\"passkey\") is stored at:\n  %s\n", recovery)
	fmt.Println("\nKeep it private. Anyone with it can log in as you.")
}

// runLogin creates a new account (name) or restores one (user_id + passkey).
//
//	kyrc login              interactive: choose new or restore
//	kyrc login <name>       create a new account with that name
func runLogin(args []string) {
	dir, err := store.ConfigDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "kyrc:", err)
		os.Exit(1)
	}

	if identity.Exists(dir) {
		name, userID, _, _ := identity.LoadPublic(dir)
		fmt.Printf("Already logged in as %s (%s).\n", name, userID)
		fmt.Println("Delete the kyrc config dir to switch accounts, or run `kyrc whoami`.")
		return
	}

	reader := bufio.NewReader(os.Stdin)

	// Non-interactive fast path: `kyrc login Alice`.
	if len(args) > 0 {
		createAccount(dir, strings.TrimSpace(strings.Join(args, " ")))
		return
	}

	fmt.Println("Welcome to kyrc. Set up your leaderboard identity:")
	fmt.Println("  [1] Create a new account")
	fmt.Println("  [2] Restore an existing account (user_id + passkey)")
	fmt.Print("Choose 1 or 2: ")
	choice, _ := reader.ReadString('\n')
	switch strings.TrimSpace(choice) {
	case "2":
		fmt.Print("Enter your name: ")
		name, _ := reader.ReadString('\n')
		fmt.Print("Enter your recovery phrase (passkey): ")
		phrase, _ := reader.ReadString('\n')
		id, err := identity.Restore(strings.TrimSpace(name), phrase)
		if err != nil {
			fmt.Fprintln(os.Stderr, "kyrc: restore failed:", err)
			os.Exit(1)
		}
		if err := id.Save(dir); err != nil {
			fmt.Fprintln(os.Stderr, "kyrc:", err)
			os.Exit(1)
		}
		fmt.Printf("\nRestored account %s (%s).\n", id.Name, id.UserID)
	default:
		fmt.Print("Enter a display name: ")
		name, _ := reader.ReadString('\n')
		createAccount(dir, strings.TrimSpace(name))
	}
}

func createAccount(dir, name string) {
	id, err := identity.New(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, "kyrc:", err)
		os.Exit(1)
	}
	if err := id.Save(dir); err != nil {
		fmt.Fprintln(os.Stderr, "kyrc:", err)
		os.Exit(1)
	}
	_, _, recovery := identity.Paths(dir)
	fmt.Printf("\n✓ Account created.\n\n")
	fmt.Printf("  name:     %s\n", id.Name)
	fmt.Printf("  user_id:  %s\n", id.UserID)
	fmt.Printf("  passkey:  %s\n", id.EncodePasskey())
	fmt.Printf("\nWrite the passkey down. It is the ONLY way to log in on another\n")
	fmt.Printf("machine. It's also saved (privately) at:\n  %s\n", recovery)
}

// runLeaderboard fetches and prints the online leaderboard.
func runLeaderboard() {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	entries, err := leaderboard.NewClient(baseURL()).Top(ctx, 20)
	if err != nil {
		fmt.Fprintln(os.Stderr, "kyrc: could not reach the leaderboard (are you online?):", err)
		os.Exit(1)
	}
	if len(entries) == 0 {
		fmt.Println("The leaderboard is empty — be the first to submit a score!")
		return
	}
	fmt.Printf("%-4s %-20s %8s %6s\n", "#", "name", "wpm", "acc")
	for _, e := range entries {
		fmt.Printf("%-4d %-20s %8.1f %5.0f%%\n", e.Rank, truncate(e.Name, 20), e.WPM, e.Accuracy*100)
	}
}

// runSync forces a leaderboard push of the user's best result.
func runSync() {
	dir, err := store.ConfigDir()
	if err != nil || !identity.Exists(dir) {
		fmt.Println("Not logged in. Run `kyrc login` first.")
		return
	}
	id, err := identity.Load(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "kyrc:", err)
		os.Exit(1)
	}
	path, _ := store.DefaultPath()
	h, err := store.Load(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "kyrc:", err)
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := leaderboard.NewClient(baseURL()).SyncBest(ctx, id, h); err != nil {
		fmt.Fprintln(os.Stderr, "kyrc: sync failed:", err)
		os.Exit(1)
	}
	fmt.Println("✓ Synced your best result to the leaderboard.")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
