package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// kyrc is installed through many package managers, each with its own update
// path. Overwriting a package-manager-managed binary would fight that manager
// (e.g. clobbering a Homebrew binary breaks `brew`, and npm re-overwrites on
// the next update). So `kyrc update` does the safe, correct thing: it checks
// the latest release, and if you're behind, detects HOW kyrc was installed and
// tells you the exact command for that installer. It never silently swaps the
// binary out from under a package manager.

const latestReleaseAPI = "https://api.github.com/repos/abh1nav9/kyrc/releases/latest"

// runUpdate implements `kyrc update` (and `kyrc update --check`).
func runUpdate(args []string) {
	checkOnly := false
	for _, a := range args {
		if a == "--check" {
			checkOnly = true
		}
	}

	fmt.Println("Checking for updates…")
	latest, err := latestVersion()
	if err != nil {
		fmt.Fprintln(os.Stderr, "kyrc: could not check latest version:", err)
		fmt.Fprintln(os.Stderr, "Are you online? You can also check https://github.com/abh1nav9/kyrc/releases")
		os.Exit(1)
	}

	current := normalizeVersion(version)
	fmt.Printf("  installed: %s\n", displayVersion(version))
	fmt.Printf("  latest:    %s\n", latest)

	if current != "" && current == normalizeVersion(latest) {
		fmt.Println("\n✓ You're on the latest version.")
		return
	}
	if current == "" {
		// Dev/unknown build — can't compare, so just show how to update.
		fmt.Println("\n(dev build — showing how to update to the latest release)")
	} else {
		fmt.Printf("\nA newer version is available: %s → %s\n", displayVersion(version), latest)
	}

	if checkOnly {
		return
	}

	method, cmd := updateCommand()
	fmt.Printf("\nkyrc was installed via %s. To update, run:\n\n", method)
	fmt.Printf("  %s\n", cmd)
}

// latestVersion queries the GitHub Releases API for the latest tag.
func latestVersion() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestReleaseAPI, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github api status %d", resp.StatusCode)
	}
	var out struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.TagName == "" {
		return "", fmt.Errorf("no release found")
	}
	return out.TagName, nil
}

// updateCommand detects the install method from the running binary's path and
// returns a human label plus the exact command to update.
func updateCommand() (method, cmd string) {
	exe, err := os.Executable()
	if err == nil {
		if resolved, e := filepath.EvalSymlinks(exe); e == nil {
			exe = resolved
		}
	}
	p := strings.ToLower(exe)

	switch {
	case strings.Contains(p, "node_modules") || strings.Contains(p, "npm"):
		return "npm", "npm update -g @kyrc/kyrc"
	case strings.Contains(p, "cellar") || strings.Contains(p, "homebrew"):
		return "Homebrew", "brew upgrade kyrc"
	case strings.Contains(p, "scoop"):
		return "Scoop", "scoop update kyrc"
	case strings.Contains(p, "winget") || strings.Contains(p, "windowsapps"):
		return "WinGet", "winget upgrade abh1nav9.kyrc"
	case strings.Contains(p, "/snap/") || strings.Contains(p, "snap"):
		return "Snap", "sudo snap refresh kyrc"
	case strings.HasPrefix(p, "/usr/bin") || strings.HasPrefix(p, "/usr/local/bin"):
		// Installed from a .deb/.rpm or a raw binary drop. Prefer the repo
		// instructions since we can't tell apt vs dnf vs manual reliably.
		return "a system package (apt/dnf) or manual install",
			"apt update && apt install --only-upgrade kyrc   # (or: dnf upgrade kyrc)"
	default:
		return "a direct download",
			"grab the latest static binary from https://github.com/abh1nav9/kyrc/releases/latest"
	}
}

// normalizeVersion strips a leading "v" and returns "" for non-release builds
// (dev/none/unknown) so we don't false-compare.
func normalizeVersion(v string) string {
	v = strings.TrimSpace(strings.TrimPrefix(v, "v"))
	switch v {
	case "", "dev", "none", "unknown":
		return ""
	}
	return v
}

func displayVersion(v string) string {
	if normalizeVersion(v) == "" {
		return v + " (unreleased)"
	}
	return v
}
