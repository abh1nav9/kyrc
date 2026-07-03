package ui

import "github.com/mattn/go-runewidth"

// lineBreaks computes indices into the target rune slice where a newline
// should be inserted, wrapping on spaces at the given display width. It
// works on the PLAIN target (no ANSI), using display width (CJK/emoji are
// two cells), so wrapping stays correct regardless of styling applied
// later. Returns a set of rune indices *before which* to break.
func lineBreaks(target []rune, width int) map[int]bool {
	breaks := map[int]bool{}
	if width <= 0 {
		return breaks
	}

	col := 0
	lastSpace := -1 // index of last space seen on the current line
	lineStart := 0

	for i := 0; i < len(target); i++ {
		r := target[i]
		if r == ' ' {
			lastSpace = i
		}
		col += runewidth.RuneWidth(r)

		if col > width {
			if lastSpace > lineStart {
				// Break after the last space: the next word starts fresh.
				breaks[lastSpace+1] = true
				lineStart = lastSpace + 1
				// Recompute column for the carried-over partial word.
				col = 0
				for j := lastSpace + 1; j <= i; j++ {
					col += runewidth.RuneWidth(target[j])
				}
				lastSpace = -1
			} else {
				// A single word longer than the line: hard break here.
				breaks[i] = true
				lineStart = i
				col = runewidth.RuneWidth(r)
				lastSpace = -1
			}
		}
	}
	return breaks
}
