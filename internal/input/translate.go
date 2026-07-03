// Package input translates terminal key messages into engine events.
//
// Bubble Tea owns raw mode, the read loop, UTF-8 decoding, and escape-
// sequence disambiguation (the ESC-timeout problem). Our job is the thin
// but critical step of stamping each key with a capture time and mapping
// it to a domain event. The timestamp is taken here — the closest point
// to input we control — so metrics inherit as little scheduler jitter as
// possible.
package input

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/abh1nav9/kyrc/internal/engine"
)

// Action is the decoded intent of a keypress that the UI must act on.
type Action int

const (
	// ActionNone: key is irrelevant to typing (ignored).
	ActionNone Action = iota
	// ActionEvent: a producing/deletion event to feed the engine.
	ActionEvent
	// ActionQuit: user asked to exit.
	ActionQuit
	// ActionRestart: user asked to restart the test (tab / ctrl+r).
	ActionRestart
	// ActionPasteRejected: a paste was detected and refused. The UI
	// should flash a warning; the paste is NOT fed to the engine.
	ActionPasteRejected
)

// Decoded is the result of translating one key message.
type Decoded struct {
	Action Action
	Event  engine.Event
}

// Translate maps a Bubble Tea key message to a Decoded action, stamping
// the event with `now` (inject time.Now() at the call site, or a fixed
// clock in tests). Paste is detected via the multi-rune Runes payload
// that bracketed-paste delivers as a single KeyRunes message.
func Translate(msg tea.KeyMsg, now time.Time) Decoded {
	switch msg.Type {
	case tea.KeyCtrlC:
		return Decoded{Action: ActionQuit}
	case tea.KeyEsc:
		return Decoded{Action: ActionQuit}
	case tea.KeyTab:
		return Decoded{Action: ActionRestart}
	case tea.KeyCtrlR:
		return Decoded{Action: ActionRestart}
	case tea.KeyBackspace, tea.KeyDelete:
		return Decoded{Action: ActionEvent, Event: engine.NewBackspace(now)}
	case tea.KeyCtrlW:
		return Decoded{Action: ActionEvent, Event: engine.NewWordDelete(now)}
	case tea.KeySpace:
		return Decoded{Action: ActionEvent, Event: engine.NewSpace(now)}
	case tea.KeyRunes:
		// A single rune is a normal keystroke. Multiple runes in one
		// message means a paste (or an IME commit) — reject it so typed
		// bursts can't inflate WPM, exactly as Monkeytype does.
		if len(msg.Runes) == 1 {
			r := msg.Runes[0]
			if r == ' ' {
				return Decoded{Action: ActionEvent, Event: engine.NewSpace(now)}
			}
			return Decoded{Action: ActionEvent, Event: engine.NewRune(r, now)}
		}
		if len(msg.Runes) > 1 {
			return Decoded{Action: ActionPasteRejected}
		}
	}
	return Decoded{Action: ActionNone}
}
