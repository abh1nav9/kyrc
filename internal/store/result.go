// Package store persists typing-test results locally, fully offline.
//
// kyrc keeps the last N results in a single JSON file under the user's
// config directory. A Result is self-contained: it carries the derived
// metrics for display AND the raw keystroke log, so a result can be
// re-scored (audited) or later submitted to the leaderboard where the
// server replays the log to verify the WPM. Nothing here touches the
// network.
package store

import (
	"time"

	"github.com/abh1nav9/kyrc/internal/engine"
)

// maxResults is how many recent results we retain. Oldest are dropped.
const maxResults = 10

// Result is one completed test. Fields are a superset of engine.Metrics
// (flattened for stable JSON) plus the audit log and test parameters.
type Result struct {
	// Completed is when the test finished (wall clock, for display/sort).
	Completed time.Time `json:"completed"`

	// Mode is "words" or "time"; Param is the word count or seconds.
	Mode  string `json:"mode"`
	Param int    `json:"param"`

	// Derived metrics (mirror engine.Metrics; see that type for defs).
	WPM         float64 `json:"wpm"`
	RawWPM      float64 `json:"raw_wpm"`
	Accuracy    float64 `json:"accuracy"`
	Consistency float64 `json:"consistency"`

	// ElapsedMS is the measured duration in milliseconds.
	ElapsedMS int64 `json:"elapsed_ms"`

	// Log is the append-only keystroke record. This is what makes a result
	// auditable/replayable. It can be large-ish but N is small (<=10).
	Log []engine.Keystroke `json:"log"`
}

// FromMetrics builds a Result from a finished session's metrics + log.
func FromMetrics(mode string, param int, m engine.Metrics, log []engine.Keystroke, completed time.Time) Result {
	return Result{
		Completed:   completed,
		Mode:        mode,
		Param:       param,
		WPM:         m.WPM,
		RawWPM:      m.RawWPM,
		Accuracy:    m.Accuracy,
		Consistency: m.Consistency,
		ElapsedMS:   m.Elapsed.Milliseconds(),
		Log:         log,
	}
}
