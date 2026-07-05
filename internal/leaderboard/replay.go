package leaderboard

import (
	"time"

	"github.com/abh1nav9/kyrc/internal/engine"
)

// Replay recomputes metrics from a submission's keystroke log ALONE, deriving
// the elapsed time from the log's own timestamps rather than trusting the
// client's claimed ElapsedMS. This is the anti-cheat linchpin: a client
// cannot inflate WPM by claiming a short duration, because the duration comes
// from the timestamps embedded in (and signed with) the log.
//
// The server compares the replayed WPM against the claimed WPM and rejects
// submissions that disagree beyond a small tolerance.
func Replay(log []engine.Keystroke) engine.Metrics {
	return engine.Compute(log, elapsedFromLog(log))
}

// elapsedFromLog is the span from the first to the last keystroke timestamp.
// The clock starts on the first keystroke (matching the engine), so this is
// the true typing duration regardless of what the client claimed.
func elapsedFromLog(log []engine.Keystroke) time.Duration {
	if len(log) < 2 {
		return 0
	}
	first := log[0].At
	last := log[0].At
	for _, k := range log[1:] {
		if k.At.Before(first) {
			first = k.At
		}
		if k.At.After(last) {
			last = k.At
		}
	}
	return last.Sub(first)
}

// WPMTolerance is how far a claimed WPM may differ from the replayed WPM
// before the server rejects the submission. Small rounding drift is fine;
// fabrication is not.
const WPMTolerance = 1.0

// Accept reports whether a submission's claimed WPM matches its replayed WPM
// within tolerance, and returns the authoritative replayed metrics. The
// server stores the REPLAYED metrics, never the claimed ones.
func Accept(s Submission) (engine.Metrics, bool) {
	m := Replay(s.Log)
	diff := m.WPM - s.WPM
	if diff < 0 {
		diff = -diff
	}
	return m, diff <= WPMTolerance
}
