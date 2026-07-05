package store

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// History is an in-memory view of the persisted results, newest-appended.
// It is not safe for concurrent use; the CLI is single-goroutine for this.
type History struct {
	path    string
	Results []Result
}

// Sort orders results by WPM.
type Sort int

const (
	// SortRecent is chronological, newest first (default view).
	SortRecent Sort = iota
	// SortTopWPM is highest WPM first.
	SortTopWPM
	// SortLowWPM is lowest WPM first.
	SortLowWPM
)

// DefaultPath returns the results file location under the OS config dir,
// e.g. ~/.config/kyrc/results.json (Linux) or the platform equivalent.
// It does not create anything.
func DefaultPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "results.json"), nil
}

// configDir returns kyrc's config directory (created lazily by callers
// that write). Uses os.UserConfigDir so it's correct on every platform.
func configDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "kyrc"), nil
}

// Load reads the results file at path. A missing file is not an error —
// it yields an empty History ready to append to.
func Load(path string) (*History, error) {
	h := &History{path: path}
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return h, nil
		}
		return nil, err
	}
	if len(b) == 0 {
		return h, nil
	}
	if err := json.Unmarshal(b, &h.Results); err != nil {
		// A corrupt file shouldn't brick the app. Start fresh but keep a
		// backup so nothing is silently destroyed.
		_ = os.Rename(path, path+".corrupt")
		h.Results = nil
		return h, nil
	}
	return h, nil
}

// Add appends a result, trims to the most recent maxResults, and persists
// atomically. The newest result is always kept.
func (h *History) Add(r Result) error {
	h.Results = append(h.Results, r)
	if len(h.Results) > maxResults {
		h.Results = h.Results[len(h.Results)-maxResults:]
	}
	return h.save()
}

// Sorted returns a copy of the results in the requested order. The stored
// slice is never mutated, so the on-disk order (chronological) is stable.
func (h *History) Sorted(s Sort) []Result {
	out := make([]Result, len(h.Results))
	copy(out, h.Results)
	switch s {
	case SortTopWPM:
		sort.SliceStable(out, func(i, j int) bool { return out[i].WPM > out[j].WPM })
	case SortLowWPM:
		sort.SliceStable(out, func(i, j int) bool { return out[i].WPM < out[j].WPM })
	case SortRecent:
		sort.SliceStable(out, func(i, j int) bool { return out[i].Completed.After(out[j].Completed) })
	}
	return out
}

// Best returns the highest-WPM result and true, or a zero Result and false
// when there is no history. Handy for "personal best" display and for
// deciding what to push to the leaderboard.
func (h *History) Best() (Result, bool) {
	if len(h.Results) == 0 {
		return Result{}, false
	}
	best := h.Results[0]
	for _, r := range h.Results[1:] {
		if r.WPM > best.WPM {
			best = r
		}
	}
	return best, true
}

// save writes the file atomically (temp + rename) so a crash mid-write can
// never leave a half-written results file.
func (h *History) save() error {
	dir := filepath.Dir(h.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(h.Results, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "results-*.json.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op after successful rename
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, h.path)
}
