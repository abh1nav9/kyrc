package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/abh1nav9/kyrc/internal/engine"
)

// Palette. Styles are package-level so they're built once, not per frame —
// the hot path (View on every keystroke) must stay allocation-light.
var (
	styleUntyped = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // dim grey
	styleCorrect = lipgloss.NewStyle().Foreground(lipgloss.Color("252")) // near-white
	styleWrong   = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Underline(true)
	styleCaret   = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("220"))
	styleExtra   = lipgloss.NewStyle().Foreground(lipgloss.Color("124"))

	styleStat      = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	styleStatLabel = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	styleHint      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	styleWarn      = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("203")).Bold(true)
	styleTitle     = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Bold(true)
)

// View renders the whole screen as a pure function of the model. No state
// mutation happens here.
func (m Model) View() string {
	if m.quitting {
		return ""
	}
	if m.session.Phase() == engine.PhaseFinished {
		return m.viewResults()
	}
	return m.viewTest()
}

func (m Model) viewTest() string {
	var b strings.Builder

	// Live clock line: countdown for timed mode, elapsed otherwise.
	b.WriteString(m.viewClockLine())
	b.WriteString("\n\n")

	b.WriteString(m.renderTarget())
	b.WriteString("\n\n")

	if m.now.Before(m.warnUntil) {
		b.WriteString(styleWarn.Render(" paste ignored — type it out "))
		b.WriteString("\n\n")
	}

	b.WriteString(styleHint.Render("tab restart   ·   esc quit"))
	return m.center(b.String())
}

func (m Model) viewClockLine() string {
	if m.cfg.Mode == engine.ModeTime {
		rem := m.session.Remaining(m.now)
		secs := int(rem.Seconds() + 0.999) // ceil so it hits 0 only at end
		return styleStat.Render(fmt.Sprintf("%ds", secs))
	}
	el := m.session.Elapsed(m.now)
	return styleStat.Render(fmt.Sprintf("%.1fs", el.Seconds()))
}

// renderTarget builds the color-coded passage with a caret at the cursor.
// It walks the target and the typed runes together, classifying each cell.
func (m Model) renderTarget() string {
	target := m.session.Target()
	typed := m.session.Typed()
	cursor := m.session.Cursor()

	breaks := lineBreaks(target, m.wrapWidth())

	var b strings.Builder
	for i, tr := range target {
		if breaks[i] {
			b.WriteByte('\n')
		}
		switch {
		case i < len(typed):
			if typed[i] == tr {
				b.WriteString(styleCorrect.Render(string(tr)))
			} else {
				// Show the intended character but marked wrong, so the
				// user reads the passage, not their mistakes. Spaces get
				// a visible marker when mistyped.
				ch := tr
				if tr == ' ' {
					ch = '_'
				}
				b.WriteString(styleWrong.Render(string(ch)))
			}
		case i == cursor:
			b.WriteString(styleCaret.Render(string(tr)))
		default:
			b.WriteString(styleUntyped.Render(string(tr)))
		}
	}

	// Extra characters typed past the end of the target (over-typing).
	if len(typed) > len(target) {
		for _, r := range typed[len(target):] {
			b.WriteString(styleExtra.Render(string(r)))
		}
	}
	return b.String()
}

func (m Model) viewResults() string {
	mtr := engine.Compute(m.session.Log(), m.session.Elapsed(m.now))

	stat := func(label, val string) string {
		return styleStat.Render(val) + "  " + styleStatLabel.Render(label)
	}

	rows := []string{
		styleTitle.Render("results"),
		"",
		stat("wpm", fmt.Sprintf("%.0f", mtr.WPM)),
		stat("raw", fmt.Sprintf("%.0f", mtr.RawWPM)),
		stat("acc", fmt.Sprintf("%.0f%%", mtr.Accuracy*100)),
		stat("consistency", fmt.Sprintf("%.0f%%", mtr.Consistency*100)),
		stat("time", fmt.Sprintf("%.1fs", mtr.Elapsed.Seconds())),
		stat("chars", fmt.Sprintf("%d/%d", mtr.CorrectChars, mtr.TypedChars)),
		"",
		styleHint.Render("tab restart   ·   esc quit"),
	}
	return m.center(strings.Join(rows, "\n"))
}

// center places content in the middle of the terminal when dimensions are
// known; otherwise returns it as-is (e.g. before the first size message).
func (m Model) center(s string) string {
	if m.width == 0 || m.height == 0 {
		return s
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, s)
}

func (m Model) wrapWidth() int {
	if m.width == 0 {
		return 60
	}
	w := m.width - 8
	if w > 80 {
		w = 80
	}
	if w < 20 {
		w = 20
	}
	return w
}
