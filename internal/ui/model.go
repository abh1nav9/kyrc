// Package ui is the Bubble Tea adapter. It owns nothing about WPM math or
// the typing state machine — it observes an engine.Session and renders it.
// This hard boundary keeps the engine unit-testable without a terminal and
// lets the same core back a web or benchmark front-end later.
package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/abh1nav9/kyrc/internal/engine"
	"github.com/abh1nav9/kyrc/internal/input"
	"github.com/abh1nav9/kyrc/internal/wordsource"
)

// Config controls how a test is built.
type Config struct {
	Mode      engine.Mode
	WordCount int           // for ModeWords
	Duration  time.Duration // for ModeTime
	Source    wordsource.Source
}

// tickMsg drives the clock display. It is intentionally SEPARATE from
// keystroke handling: typing feedback is immediate (per key), while the
// countdown/elapsed readout only needs to refresh a few times a second.
// A late clock frame is invisible; a late keystroke is not.
type tickMsg time.Time

// Model is the Bubble Tea model. It holds a pointer to the source-of-truth
// session plus pure view state (warnings, dimensions).
type Model struct {
	cfg     Config
	session *engine.Session

	width  int
	height int

	// warnUntil suppresses/expires the transient paste warning banner.
	warnUntil time.Time
	now       time.Time // last observed time, for rendering only

	quitting bool

	// OnFinish, if set, is called once when a test transitions to Finished.
	// It receives the completed session so the caller can persist results
	// and (best-effort) sync — the UI stays ignorant of storage/network.
	// The boolean guards against calling it more than once per test.
	OnFinish func(*engine.Session, Config)
	finished bool
}

// New builds a Model with a freshly generated test.
func New(cfg Config) Model {
	return Model{cfg: cfg, session: buildSession(cfg), now: time.Now()}
}

func buildSession(cfg Config) *engine.Session {
	switch cfg.Mode {
	case engine.ModeTime:
		// Generate a generous stream so a fast typist won't exhaust it.
		target := cfg.Source.Words(maxInt(cfg.WordCount, 200))
		return engine.NewTimed(target, cfg.Duration)
	default:
		return engine.NewWords(cfg.Source.Words(cfg.WordCount))
	}
}

// Init starts the clock ticker.
func (m Model) Init() tea.Cmd {
	return tick()
}

func tick() tea.Cmd {
	// ~15Hz clock refresh: smooth enough for a countdown, cheap on CPU.
	return tea.Tick(66*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update is the sole state transition. It captures the timestamp once, up
// front, so the engine sees a single coherent clock reading per message.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	now := time.Now()
	m.now = now

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tickMsg:
		m.session.Tick(now)
		if m.session.Phase() == engine.PhaseFinished {
			m.fireFinish()
			return m, nil // stop ticking; results screen is static
		}
		return m, tick()

	case tea.KeyMsg:
		return m.handleKey(msg, now)
	}
	return m, nil
}

// fireFinish invokes OnFinish exactly once per completed test.
func (m *Model) fireFinish() {
	if m.finished || m.OnFinish == nil {
		return
	}
	m.finished = true
	m.OnFinish(m.session, m.cfg)
}

func (m Model) handleKey(msg tea.KeyMsg, now time.Time) (tea.Model, tea.Cmd) {
	// On the results screen, only quit/restart are live.
	if m.session.Phase() == engine.PhaseFinished {
		d := input.Translate(msg, now)
		switch d.Action {
		case input.ActionQuit:
			m.quitting = true
			return m, tea.Quit
		case input.ActionRestart:
			return m.restart(), tick()
		}
		return m, nil
	}

	d := input.Translate(msg, now)
	switch d.Action {
	case input.ActionQuit:
		m.quitting = true
		return m, tea.Quit
	case input.ActionRestart:
		return m.restart(), tick()
	case input.ActionPasteRejected:
		m.warnUntil = now.Add(1500 * time.Millisecond)
		return m, nil
	case input.ActionEvent:
		m.session.Apply(d.Event)
		if m.session.Phase() == engine.PhaseFinished {
			m.fireFinish() // word-mode tests finish on the last keystroke
		}
		return m, nil
	}
	return m, nil
}

func (m Model) restart() Model {
	m.session = buildSession(m.cfg)
	m.warnUntil = time.Time{}
	m.finished = false
	return m
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
