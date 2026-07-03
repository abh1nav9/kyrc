package ui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/abh1nav9/kyrc/internal/engine"
	"github.com/abh1nav9/kyrc/internal/wordsource"
)

// feedKey drives a KeyMsg through Update and returns the new model.
func feedKey(m Model, msg tea.KeyMsg) Model {
	nm, _ := m.Update(msg)
	return nm.(Model)
}

func runes(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func TestTypingDrivesSessionToFinish(t *testing.T) {
	cfg := Config{Mode: engine.ModeWords, WordCount: 0,
		Source: wordsource.Static{Text: "ab"}}
	m := New(cfg)

	m = feedKey(m, runes("a"))
	if m.session.Phase() != engine.PhaseRunning {
		t.Fatalf("should be running after first key")
	}
	m = feedKey(m, runes("b"))
	if m.session.Phase() != engine.PhaseFinished {
		t.Fatalf("should finish when target complete, got %v", m.session.Phase())
	}
	// Results view must render without panic and contain wpm label.
	if out := m.View(); out == "" {
		t.Fatal("results view empty")
	}
}

func TestPasteIsRejected(t *testing.T) {
	cfg := Config{Mode: engine.ModeWords,
		Source: wordsource.Static{Text: "hello world here"}}
	m := New(cfg)
	m = feedKey(m, runes("pasted whole thing")) // multi-rune => paste
	if m.session.Cursor() != 0 {
		t.Fatalf("paste must not advance cursor, got %d", m.session.Cursor())
	}
	if m.now.After(m.warnUntil) {
		t.Fatal("paste should arm the warning banner")
	}
}

func TestRestartResetsSession(t *testing.T) {
	cfg := Config{Mode: engine.ModeWords,
		Source: wordsource.Static{Text: "ab"}}
	m := New(cfg)
	m = feedKey(m, runes("a"))
	m = feedKey(m, tea.KeyMsg{Type: tea.KeyTab})
	if m.session.Phase() != engine.PhaseIdle {
		t.Fatalf("restart must return to idle, got %v", m.session.Phase())
	}
	if m.session.Cursor() != 0 {
		t.Fatal("restart must clear typed input")
	}
}

func TestQuitOnEsc(t *testing.T) {
	m := New(Config{Mode: engine.ModeWords, Source: wordsource.Static{Text: "ab"}})
	nm, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !nm.(Model).quitting {
		t.Fatal("esc should set quitting")
	}
	if cmd == nil {
		t.Fatal("esc should return tea.Quit command")
	}
}

func TestViewRendersBeforeSizeMsg(t *testing.T) {
	// Rendering must not panic when width/height are still zero.
	m := New(Config{Mode: engine.ModeWords, Source: wordsource.Static{Text: "hello world"}})
	_ = m.View()
	m, _ = mustModel(m.Update(tea.WindowSizeMsg{Width: 80, Height: 24}))
	_ = m.View()
}

func mustModel(tm tea.Model, _ tea.Cmd) (Model, tea.Cmd) {
	return tm.(Model), nil
}

func TestTimedTickFinishes(t *testing.T) {
	m := New(Config{Mode: engine.ModeTime, Duration: 50 * time.Millisecond,
		WordCount: 10, Source: wordsource.NewRandom(nil)})
	m = feedKey(m, runes("a")) // start clock
	time.Sleep(60 * time.Millisecond)
	nm, _ := m.Update(tickMsg(time.Now()))
	if nm.(Model).session.Phase() != engine.PhaseFinished {
		t.Fatalf("timed test should finish after deadline tick")
	}
}
