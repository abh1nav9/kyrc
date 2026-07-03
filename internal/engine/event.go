package engine

import "time"

// KeyKind classifies a logical keystroke fed into the engine.
// The input layer is responsible for translating raw terminal bytes
// (escape sequences, UTF-8 runes, control codes) into these kinds.
type KeyKind int

const (
	// KeyRune is a printable character (a grapheme / rune).
	KeyRune KeyKind = iota
	// KeyBackspace deletes the character to the left of the cursor.
	KeyBackspace
	// KeyWordDelete deletes the previous word (ctrl+w / ctrl+backspace).
	KeyWordDelete
	// KeySpace is treated specially: it advances between words.
	KeySpace
)

// Event is a single timestamped input, captured as close to the read
// syscall as possible. The engine is a pure function of the ordered
// stream of events: replaying the same events yields identical state
// and identical metrics. Never sample the clock inside the engine —
// time always arrives with the event.
type Event struct {
	Kind KeyKind
	// Rune is only meaningful when Kind == KeyRune.
	Rune rune
	// At is the moment the key was captured. This is the ONLY source of
	// time the engine sees; it must be injected, never read internally.
	At time.Time
}

// NewRune builds a printable-character event.
func NewRune(r rune, at time.Time) Event { return Event{Kind: KeyRune, Rune: r, At: at} }

// NewSpace builds a space event.
func NewSpace(at time.Time) Event { return Event{Kind: KeySpace, At: at} }

// NewBackspace builds a backspace event.
func NewBackspace(at time.Time) Event { return Event{Kind: KeyBackspace, At: at} }

// NewWordDelete builds a word-delete event.
func NewWordDelete(at time.Time) Event { return Event{Kind: KeyWordDelete, At: at} }
