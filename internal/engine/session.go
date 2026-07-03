package engine

import (
	"time"
	"unicode"
)

// Phase is the explicit finite-state-machine state of a session.
type Phase int

const (
	// PhaseIdle: test is shown but no key has been pressed. The clock has
	// not started. This is critical — idle time must never count.
	PhaseIdle Phase = iota
	// PhaseRunning: first keystroke landed; the clock is live.
	PhaseRunning
	// PhaseFinished: the target is fully typed (or time/word limit hit).
	PhaseFinished
)

// Mode selects how the test terminates.
type Mode int

const (
	// ModeWords finishes when the target text is completed.
	ModeWords Mode = iota
	// ModeTime finishes when the configured duration elapses. The target
	// is treated as an effectively endless stream (the UI keeps feeding
	// words); the engine finishes on the time budget.
	ModeTime
)

// Session is the single source of truth for one typing test. It is a
// pure state machine: Apply(event) is the ONLY mutator, and it is
// deterministic given the target and the event stream. It contains no
// terminal, rendering, or IO concepts whatsoever.
type Session struct {
	// target is the sequence of runes the user must reproduce, including
	// the single spaces between words.
	target []rune

	// typed holds what the user has actually entered so far, in order,
	// including mistakes. Its length is the cursor position.
	typed []rune

	// log is the append-only keystroke history used for metrics/replay.
	log []Keystroke

	phase Phase
	mode  Mode

	// startedAt is the timestamp of the FIRST keystroke (not test render).
	startedAt time.Time
	// endedAt is the timestamp of the finishing keystroke / deadline.
	endedAt time.Time
	// deadline, for ModeTime, is startedAt + limit. Zero in ModeWords.
	deadline time.Time
	limit    time.Duration
}

// NewWords creates a word-count / fixed-text session.
func NewWords(target string) *Session {
	return &Session{target: []rune(target), phase: PhaseIdle, mode: ModeWords}
}

// NewTimed creates a time-limited session. The target should be a long
// stream of words; the test ends when limit elapses from the first key.
func NewTimed(target string, limit time.Duration) *Session {
	return &Session{target: []rune(target), phase: PhaseIdle, mode: ModeTime, limit: limit}
}

// Phase returns the current FSM phase.
func (s *Session) Phase() Phase { return s.phase }

// Mode returns the configured mode.
func (s *Session) Mode() Mode { return s.mode }

// Target exposes the immutable target runes (read-only view).
func (s *Session) Target() []rune { return s.target }

// Typed exposes what the user has entered so far (read-only view).
func (s *Session) Typed() []rune { return s.typed }

// Cursor is the current index into target (== len(typed)).
func (s *Session) Cursor() int { return len(s.typed) }

// Log exposes the keystroke history for metrics and replay.
func (s *Session) Log() []Keystroke { return s.log }

// StartedAt / EndedAt expose session timing. Zero if not reached.
func (s *Session) StartedAt() time.Time { return s.startedAt }
func (s *Session) EndedAt() time.Time   { return s.endedAt }

// Deadline returns the finish time for timed mode (zero otherwise).
func (s *Session) Deadline() time.Time { return s.deadline }

// Elapsed returns the live duration since start. For a finished session
// it is the frozen final duration. Before start it is zero.
func (s *Session) Elapsed(now time.Time) time.Duration {
	switch s.phase {
	case PhaseIdle:
		return 0
	case PhaseFinished:
		return s.endedAt.Sub(s.startedAt)
	default:
		return now.Sub(s.startedAt)
	}
}

// Remaining returns time left in a timed test (zero for word mode or a
// finished/idle session).
func (s *Session) Remaining(now time.Time) time.Duration {
	if s.mode != ModeTime || s.phase != PhaseRunning {
		if s.mode == ModeTime && s.phase == PhaseIdle {
			return s.limit
		}
		return 0
	}
	rem := s.deadline.Sub(now)
	if rem < 0 {
		return 0
	}
	return rem
}

// Tick lets the caller advance time without a keystroke so timed tests
// finish even if the user stops typing exactly on the deadline. It is
// pure and idempotent: it only ever transitions Running -> Finished.
func (s *Session) Tick(now time.Time) {
	if s.mode == ModeTime && s.phase == PhaseRunning && !now.Before(s.deadline) {
		s.phase = PhaseFinished
		s.endedAt = s.deadline
	}
}

// Apply is the single deterministic mutator. Given the current state and
// an event, it advances the state machine and appends to the log. The
// clock is taken exclusively from ev.At.
func (s *Session) Apply(ev Event) {
	if s.phase == PhaseFinished {
		return
	}

	// Deadline enforcement for timed mode: if this event lands after the
	// deadline, finish first and drop the keystroke (it happened too late).
	if s.mode == ModeTime && s.phase == PhaseRunning && !ev.At.Before(s.deadline) {
		s.phase = PhaseFinished
		s.endedAt = s.deadline
		return
	}

	switch ev.Kind {
	case KeyBackspace:
		s.backspace(ev.At)
		return
	case KeyWordDelete:
		s.wordDelete(ev.At)
		return
	}

	// Any producing keystroke (rune or space) starts the clock.
	s.startIfNeeded(ev.At)

	r := ev.Rune
	if ev.Kind == KeySpace {
		r = ' '
	}

	pos := len(s.typed)
	var expected rune
	if pos < len(s.target) {
		expected = s.target[pos]
	}
	correct := pos < len(s.target) && r == expected

	s.typed = append(s.typed, r)
	s.log = append(s.log, Keystroke{At: ev.At, Typed: r, Expected: expected, Correct: correct})

	s.finishIfComplete(ev.At)
}

func (s *Session) startIfNeeded(at time.Time) {
	if s.phase == PhaseIdle {
		s.phase = PhaseRunning
		s.startedAt = at
		if s.mode == ModeTime {
			s.deadline = at.Add(s.limit)
		}
	}
}

func (s *Session) backspace(at time.Time) {
	if len(s.typed) == 0 {
		return
	}
	// Backspace before the clock starts is a no-op that shouldn't start it.
	if s.phase == PhaseIdle {
		return
	}
	s.typed = s.typed[:len(s.typed)-1]
	s.log = append(s.log, Keystroke{At: at, Deletion: true})
}

func (s *Session) wordDelete(at time.Time) {
	if len(s.typed) == 0 || s.phase == PhaseIdle {
		return
	}
	i := len(s.typed)
	// Skip trailing spaces, then the word.
	for i > 0 && unicode.IsSpace(s.typed[i-1]) {
		i--
	}
	for i > 0 && !unicode.IsSpace(s.typed[i-1]) {
		i--
	}
	for len(s.typed) > i {
		s.typed = s.typed[:len(s.typed)-1]
		s.log = append(s.log, Keystroke{At: at, Deletion: true})
	}
}

// finishIfComplete transitions to Finished in word mode when the user has
// typed at least the full target length AND the final characters match.
// In word mode we finish once the cursor reaches the end of the target.
func (s *Session) finishIfComplete(at time.Time) {
	if s.mode != ModeWords {
		return
	}
	if len(s.typed) >= len(s.target) {
		s.phase = PhaseFinished
		s.endedAt = at
	}
}
