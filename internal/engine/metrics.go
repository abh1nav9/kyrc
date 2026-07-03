package engine

import (
	"math"
	"time"
)

// The canonical typing-test convention: 1 "word" == 5 characters,
// spaces included. All WPM figures derive from this.
const charsPerWord = 5.0

// Metrics is a snapshot of derived statistics. Every field is a pure
// function of the keystroke log and the elapsed duration — nothing here
// reads a clock or mutates state, so the same log always yields the same
// Metrics. That reproducibility is the whole point.
type Metrics struct {
	// WPM (net): correct characters / 5 / minutes. This is the hero number,
	// matching Monkeytype's headline "wpm".
	WPM float64
	// RawWPM: all produced characters / 5 / minutes, ignoring correctness.
	RawWPM float64
	// Accuracy: correct keystrokes / total producing keystrokes, in [0,1].
	// Deletions are excluded from the denominator; they are corrections,
	// not attempts.
	Accuracy float64
	// Consistency: 1 - coefficient of variation of per-second raw WPM,
	// clamped to [0,1]. Rewards steady typing. 1.0 == perfectly even.
	Consistency float64

	// Raw counts for display / debugging / auditing.
	CorrectChars int
	TypedChars   int // producing keystrokes (excludes deletions)
	Errors       int
	Elapsed      time.Duration
}

// Compute derives Metrics from a session's log and an elapsed duration.
// elapsed must be measured by the caller from injected timestamps
// (Session.Elapsed), never from wall-clock inside here.
func Compute(log []Keystroke, elapsed time.Duration) Metrics {
	var correct, typed int
	for _, k := range log {
		if k.Deletion {
			continue
		}
		typed++
		if k.Correct {
			correct++
		}
	}
	errors := typed - correct

	minutes := elapsed.Minutes()
	var wpm, raw float64
	if minutes > 0 {
		wpm = float64(correct) / charsPerWord / minutes
		raw = float64(typed) / charsPerWord / minutes
	}

	acc := 1.0
	if typed > 0 {
		acc = float64(correct) / float64(typed)
	}

	return Metrics{
		WPM:          wpm,
		RawWPM:       raw,
		Accuracy:     acc,
		Consistency:  consistency(log),
		CorrectChars: correct,
		TypedChars:   typed,
		Errors:       errors,
		Elapsed:      elapsed,
	}
}

// consistency measures how even the typing pace was, via the coefficient
// of variation (stddev/mean) of raw WPM computed over 1-second buckets.
// We report 1 - CV so that higher is better, clamped to [0,1].
func consistency(log []Keystroke) float64 {
	// Bucket producing keystrokes by whole second offset from the first
	// keystroke, then compute a per-second WPM sample for each bucket.
	var first time.Time
	found := false
	for _, k := range log {
		if k.Deletion {
			continue
		}
		if !found {
			first = k.At
			found = true
		}
	}
	if !found {
		return 0
	}

	buckets := map[int]int{}
	for _, k := range log {
		if k.Deletion {
			continue
		}
		sec := int(k.At.Sub(first).Seconds())
		buckets[sec]++
	}
	if len(buckets) < 2 {
		// Not enough time span to speak of variation; treat as perfect.
		return 1
	}

	samples := make([]float64, 0, len(buckets))
	for _, count := range buckets {
		// chars/sec -> WPM: (chars/5) per second * 60 seconds.
		samples = append(samples, float64(count)/charsPerWord*60.0)
	}

	mean := 0.0
	for _, v := range samples {
		mean += v
	}
	mean /= float64(len(samples))
	if mean == 0 {
		return 0
	}

	variance := 0.0
	for _, v := range samples {
		d := v - mean
		variance += d * d
	}
	variance /= float64(len(samples))
	cv := math.Sqrt(variance) / mean

	c := 1 - cv
	if c < 0 {
		c = 0
	}
	if c > 1 {
		c = 1
	}
	return c
}
