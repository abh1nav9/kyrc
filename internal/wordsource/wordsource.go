// Package wordsource generates target text for a test. It sits behind an
// interface so the engine never learns about "random words" vs "quotes"
// vs "custom text" — adding a mode never touches the state machine.
package wordsource

import (
	"math/rand"
	"strings"
)

// Source produces a target string of at least n words.
type Source interface {
	Words(n int) string
}

// English200 is the common-words list many typing tests use. Kept small
// and embedded so the binary stays dependency- and network-free.
var english200 = strings.Fields(`the be to of and a in that have i it for not on with he as you do at
this but his by from they we say her she or an will my one all would there their what so up out if about who get which go me
when make can like time no just him know take people into year your good some could them see other than then now look only come
its over think also back after use two how our work first well way even new want because any these give day most us is are was
were been has had did said get make go know take see come think look want give use find tell ask work seem feel try leave call
good new first last long great little own other old right big high different small large next early young important few public bad
same able`)

// Random draws uniformly from a word list.
type Random struct {
	words []string
	rng   *rand.Rand
}

// NewRandom builds a random English-word source. A nil rng uses a fresh
// seeded generator; injecting one keeps tests deterministic.
func NewRandom(rng *rand.Rand) *Random {
	if rng == nil {
		rng = rand.New(rand.NewSource(rand.Int63()))
	}
	return &Random{words: english200, rng: rng}
}

// Words returns n space-separated random words.
func (r *Random) Words(n int) string {
	if n <= 0 {
		return ""
	}
	var b strings.Builder
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(r.words[r.rng.Intn(len(r.words))])
	}
	return b.String()
}

// Static wraps a fixed string (custom text / quote) as a Source.
type Static struct{ Text string }

// Words returns the wrapped text regardless of n.
func (s Static) Words(int) string { return s.Text }
