package engine

import (
	"math"
	"testing"
	"time"
)

// base is a fixed epoch so tests are fully deterministic.
var base = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// typeString feeds a target-matching string into a session using synthetic
// timestamps spaced `gap` apart, starting at base. It handles spaces.
func typeString(s *Session, text string, gap time.Duration) {
	at := base
	for _, r := range text {
		if r == ' ' {
			s.Apply(NewSpace(at))
		} else {
			s.Apply(NewRune(r, at))
		}
		at = at.Add(gap)
	}
}

func TestClockStartsOnFirstKeystroke(t *testing.T) {
	s := NewWords("hello world")
	if s.Phase() != PhaseIdle {
		t.Fatalf("want idle, got %v", s.Phase())
	}
	if got := s.Elapsed(base.Add(time.Hour)); got != 0 {
		t.Fatalf("idle session must report 0 elapsed, got %v", got)
	}
	s.Apply(NewRune('h', base.Add(500*time.Millisecond)))
	if s.Phase() != PhaseRunning {
		t.Fatalf("want running after first key")
	}
	if !s.StartedAt().Equal(base.Add(500 * time.Millisecond)) {
		t.Fatalf("start anchored to first keystroke, got %v", s.StartedAt())
	}
}

func TestPerfectRunFinishesAndScores(t *testing.T) {
	target := "the quick brown fox"
	s := NewWords(target)
	// 60ms/char => 1 char per 0.06s. 19 chars.
	typeString(s, target, 60*time.Millisecond)
	if s.Phase() != PhaseFinished {
		t.Fatalf("want finished, got %v", s.Phase())
	}
	m := Compute(s.Log(), s.Elapsed(base))
	if m.Errors != 0 {
		t.Fatalf("perfect run must have 0 errors, got %d", m.Errors)
	}
	if math.Abs(m.Accuracy-1.0) > 1e-9 {
		t.Fatalf("perfect accuracy expected, got %f", m.Accuracy)
	}
	if m.CorrectChars != len([]rune(target)) {
		t.Fatalf("correct chars mismatch: %d", m.CorrectChars)
	}
}

func TestWPMComputationIsExact(t *testing.T) {
	// 25 correct chars in 6 seconds. WPM = (25/5)/(6/60) = 5/0.1 = 50.
	target := "aaaaaaaaaaaaaaaaaaaaaaaaa" // 25 a's
	s := NewWords(target)
	at := base
	for range target {
		s.Apply(NewRune('a', at))
		at = at.Add(250 * time.Millisecond) // 25 chars * 0.25 = 6.25s... adjust
	}
	// Last keystroke at base + 24*0.25 = 6.0s. Elapsed = 6.0s.
	elapsed := s.Elapsed(base)
	if elapsed != 6*time.Second {
		t.Fatalf("expected 6s elapsed, got %v", elapsed)
	}
	m := Compute(s.Log(), elapsed)
	if math.Abs(m.WPM-50.0) > 1e-9 {
		t.Fatalf("expected 50 WPM, got %f", m.WPM)
	}
}

func TestAccuracyCountsKeystrokesNotFinalText(t *testing.T) {
	// Target longer than the fix so completion doesn't fire early.
	// Type "hex" then fix the x: h, e, x(wrong, expected 'l'), backspace, l.
	s := NewWords("hello")
	at := base
	s.Apply(NewRune('h', at))
	at = at.Add(50 * time.Millisecond)
	s.Apply(NewRune('e', at))
	at = at.Add(50 * time.Millisecond)
	s.Apply(NewRune('x', at)) // wrong, expected 'l'
	at = at.Add(50 * time.Millisecond)
	s.Apply(NewBackspace(at))
	at = at.Add(50 * time.Millisecond)
	s.Apply(NewRune('l', at)) // right

	m := Compute(s.Log(), s.Elapsed(base))
	// 4 producing keystrokes (h, e, x, l), 3 correct. acc = 3/4.
	if m.TypedChars != 4 {
		t.Fatalf("expected 4 producing keystrokes, got %d", m.TypedChars)
	}
	if m.Errors != 1 {
		t.Fatalf("expected 1 error, got %d", m.Errors)
	}
	if math.Abs(m.Accuracy-3.0/4.0) > 1e-9 {
		t.Fatalf("keystroke accuracy should be 3/4, got %f", m.Accuracy)
	}
}

func TestReplayDeterminism(t *testing.T) {
	target := "replay me exactly"
	events := []Event{}
	at := base
	for _, r := range target {
		if r == ' ' {
			events = append(events, NewSpace(at))
		} else {
			events = append(events, NewRune(r, at))
		}
		at = at.Add(40 * time.Millisecond)
	}

	run := func() Metrics {
		s := NewWords(target)
		for _, e := range events {
			s.Apply(e)
		}
		return Compute(s.Log(), s.Elapsed(base))
	}
	a, b := run(), run()
	if a != b {
		t.Fatalf("replay not deterministic:\n%+v\n%+v", a, b)
	}
}

func TestTimedModeFinishesOnDeadline(t *testing.T) {
	s := NewTimed("aaaa aaaa aaaa aaaa aaaa", 2*time.Second)
	// Type past the 2s deadline; keystrokes after deadline are dropped.
	at := base
	for i := 0; i < 40; i++ {
		s.Apply(NewRune('a', at))
		at = at.Add(100 * time.Millisecond)
		if s.Phase() == PhaseFinished {
			break
		}
	}
	if s.Phase() != PhaseFinished {
		t.Fatalf("timed test must finish at deadline")
	}
	if s.Elapsed(base) != 2*time.Second {
		t.Fatalf("timed elapsed frozen to limit, got %v", s.Elapsed(base))
	}
}

func TestTickFinishesIdleDeadline(t *testing.T) {
	s := NewTimed("aaaa aaaa", 1*time.Second)
	s.Apply(NewRune('a', base)) // start clock
	// User stops typing; a tick past deadline should finish.
	s.Tick(base.Add(2 * time.Second))
	if s.Phase() != PhaseFinished {
		t.Fatalf("tick past deadline must finish, got %v", s.Phase())
	}
}

func TestWordDelete(t *testing.T) {
	s := NewWords("hello world foo")
	typeString(s, "hello wor", 30*time.Millisecond)
	before := s.Cursor()
	s.Apply(NewWordDelete(base.Add(time.Second)))
	// Should delete "wor" and the trailing space, back to "hello ".
	if got := string(s.Typed()); got != "hello " {
		t.Fatalf("word delete wrong, got %q (was cursor %d)", got, before)
	}
}
