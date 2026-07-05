package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/abh1nav9/kyrc/internal/engine"
)

func tmpStore(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "results.json")
}

func mkResult(wpm float64, at time.Time) Result {
	return Result{Completed: at, Mode: "words", Param: 25, WPM: wpm}
}

func TestLoadMissingFileIsEmpty(t *testing.T) {
	h, err := Load(tmpStore(t))
	if err != nil {
		t.Fatalf("Load missing: %v", err)
	}
	if len(h.Results) != 0 {
		t.Fatalf("expected empty, got %d", len(h.Results))
	}
}

func TestAddCapsAtTenKeepingNewest(t *testing.T) {
	path := tmpStore(t)
	h, _ := Load(path)
	base := time.Now()
	for i := 0; i < 15; i++ {
		if err := h.Add(mkResult(float64(i), base.Add(time.Duration(i)*time.Minute))); err != nil {
			t.Fatalf("Add: %v", err)
		}
	}
	if len(h.Results) != maxResults {
		t.Fatalf("expected %d results, got %d", maxResults, len(h.Results))
	}
	// The oldest (wpm 0..4) should have been dropped; newest is wpm 14.
	if h.Results[len(h.Results)-1].WPM != 14 {
		t.Fatalf("newest wpm = %v, want 14", h.Results[len(h.Results)-1].WPM)
	}
	if h.Results[0].WPM != 5 {
		t.Fatalf("oldest retained wpm = %v, want 5", h.Results[0].WPM)
	}
}

func TestPersistenceRoundTrip(t *testing.T) {
	path := tmpStore(t)
	h, _ := Load(path)
	_ = h.Add(mkResult(42, time.Now()))

	h2, err := Load(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if len(h2.Results) != 1 || h2.Results[0].WPM != 42 {
		t.Fatalf("round trip lost data: %+v", h2.Results)
	}
}

func TestSortTopAndLowWPM(t *testing.T) {
	h := &History{}
	now := time.Now()
	h.Results = []Result{mkResult(30, now), mkResult(90, now), mkResult(60, now)}

	top := h.Sorted(SortTopWPM)
	if top[0].WPM != 90 || top[2].WPM != 30 {
		t.Fatalf("SortTopWPM wrong: %v %v %v", top[0].WPM, top[1].WPM, top[2].WPM)
	}
	low := h.Sorted(SortLowWPM)
	if low[0].WPM != 30 || low[2].WPM != 90 {
		t.Fatalf("SortLowWPM wrong: %v %v %v", low[0].WPM, low[1].WPM, low[2].WPM)
	}
	// Sorting must not mutate stored order.
	if h.Results[0].WPM != 30 {
		t.Fatalf("Sorted mutated stored slice")
	}
}

func TestBest(t *testing.T) {
	h := &History{}
	if _, ok := h.Best(); ok {
		t.Fatalf("empty history should have no best")
	}
	now := time.Now()
	h.Results = []Result{mkResult(30, now), mkResult(88, now), mkResult(60, now)}
	best, ok := h.Best()
	if !ok || best.WPM != 88 {
		t.Fatalf("Best = %v (ok=%v), want 88", best.WPM, ok)
	}
}

func TestCorruptFileIsQuarantinedNotFatal(t *testing.T) {
	path := tmpStore(t)
	if err := os.WriteFile(path, []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	h, err := Load(path)
	if err != nil {
		t.Fatalf("corrupt load should not error, got %v", err)
	}
	if len(h.Results) != 0 {
		t.Fatalf("corrupt load should start empty")
	}
	if _, err := os.Stat(path + ".corrupt"); err != nil {
		t.Fatalf("expected quarantine backup, got %v", err)
	}
}

func TestFromMetricsCarriesLog(t *testing.T) {
	log := []engine.Keystroke{{Typed: 'a', Expected: 'a', Correct: true}}
	m := engine.Metrics{WPM: 50, RawWPM: 55, Accuracy: 1, Consistency: 0.9, Elapsed: 6 * time.Second}
	r := FromMetrics("words", 25, m, log, time.Now())
	if r.WPM != 50 || r.ElapsedMS != 6000 || len(r.Log) != 1 {
		t.Fatalf("FromMetrics mismatch: %+v", r)
	}
}
