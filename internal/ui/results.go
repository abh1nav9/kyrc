package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/abh1nav9/kyrc/internal/store"
)

// resultsModel is a small standalone Bubble Tea program for `kyrc results`.
// It shows the local history (last 10) and cycles sort order with `s`.
type resultsModel struct {
	h      *store.History
	sort   store.Sort
	width  int
	height int
}

// RunResults launches the history screen and blocks until the user quits.
func RunResults(h *store.History) error {
	_, err := tea.NewProgram(resultsModel{h: h, sort: store.SortRecent}, tea.WithAltScreen()).Run()
	return err
}

func (m resultsModel) Init() tea.Cmd { return nil }

func (m resultsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "s":
			// Cycle: recent → top → low → recent.
			m.sort = (m.sort + 1) % 3
		case "t":
			m.sort = store.SortTopWPM
		case "l":
			m.sort = store.SortLowWPM
		case "r":
			m.sort = store.SortRecent
		}
	}
	return m, nil
}

var (
	styleResHeader = lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Bold(true)
	styleResWPM    = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	styleResDim    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

func (m resultsModel) View() string {
	var b strings.Builder
	b.WriteString(styleTitle.Render("your last results") + "\n\n")

	rows := m.h.Sorted(m.sort)
	if len(rows) == 0 {
		b.WriteString(styleResDim.Render("No results yet — run `kyrc` to take a test.\n"))
		return m.center(b.String())
	}

	b.WriteString(styleResHeader.Render(fmt.Sprintf("%-3s %8s %6s %6s %-8s %s", "#", "wpm", "raw", "acc", "mode", "when")) + "\n")
	for i, r := range rows {
		when := r.Completed.Format("Jan 2 15:04")
		mode := fmt.Sprintf("%s %d", r.Mode, r.Param)
		line := fmt.Sprintf("%-3d %8s %6.0f %5.0f%% %-8s %s",
			i+1,
			styleResWPM.Render(fmt.Sprintf("%.1f", r.WPM)),
			r.RawWPM, r.Accuracy*100, mode, when)
		b.WriteString(line + "\n")
	}

	if best, ok := m.h.Best(); ok {
		b.WriteString("\n" + styleResDim.Render(fmt.Sprintf("personal best: %.1f wpm", best.WPM)) + "\n")
	}
	b.WriteString("\n" + styleHint.Render("sort: [s] cycle · [t] top · [l] low · [r] recent   ·   [q] quit"))
	b.WriteString("\n" + styleResDim.Render("now: "+sortName(m.sort)))

	return m.center(b.String())
}

func sortName(s store.Sort) string {
	switch s {
	case store.SortTopWPM:
		return "top wpm"
	case store.SortLowWPM:
		return "low wpm"
	default:
		return "most recent"
	}
}

func (m resultsModel) center(s string) string {
	if m.width == 0 || m.height == 0 {
		return s
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, s)
}
