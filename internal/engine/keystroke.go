package engine

import "time"

// Keystroke is an append-only record of one meaningful typing action.
// The full session is reconstructable from this log alone; every metric
// is a pure function over the slice of keystrokes. This is what lets us
// audit a result ("your WPM is wrong") by replaying the exact session.
type Keystroke struct {
	At time.Time
	// Typed is the rune the user produced. Zero for deletions.
	Typed rune
	// Expected is the rune the target text wanted at that position.
	// Zero when the user typed past the end of a word (extra chars).
	Expected rune
	// Correct reports whether Typed matched Expected at the moment it
	// was pressed. Deletions are never "correct" keystrokes; they are
	// recorded separately for consistency/undo accounting.
	Correct bool
	// Deletion marks backspace/word-delete actions. These count against
	// keystroke accuracy indirectly (via the error they correct) but are
	// not themselves scored as correct/incorrect input.
	Deletion bool
}
