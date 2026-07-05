package leaderboard

import (
	"testing"
	"time"

	"github.com/abh1nav9/kyrc/internal/engine"
	"github.com/abh1nav9/kyrc/internal/identity"
)

// buildLog fabricates a keystroke log of n correct chars over the given
// duration, evenly spaced, so Replay yields a predictable WPM.
func buildLog(n int, dur time.Duration) []engine.Keystroke {
	start := time.Unix(1_700_000_000, 0)
	log := make([]engine.Keystroke, 0, n)
	for i := 0; i < n; i++ {
		at := start
		if n > 1 {
			at = start.Add(time.Duration(i) * dur / time.Duration(n-1))
		}
		log = append(log, engine.Keystroke{At: at, Typed: 'a', Expected: 'a', Correct: true})
	}
	return log
}

func TestSubmissionSignVerify(t *testing.T) {
	id, _ := identity.New("alice")
	s := Submission{
		Version: APIVersion, UserID: id.UserID, Name: id.Name,
		Mode: "words", WPM: 50, Log: buildLog(50, 6*time.Second),
		Nonce: "abc", Timestamp: time.Now().Unix(),
	}
	SignSubmission(&s, id.Sign)

	if !VerifySubmission(id.PublicKey, s) {
		t.Fatal("valid submission failed verification")
	}
	// Tamper with WPM after signing → must fail.
	bad := s
	bad.WPM = 999
	if VerifySubmission(id.PublicKey, bad) {
		t.Fatal("tampered WPM verified")
	}
	// Another user's key must not verify.
	mallory, _ := identity.New("mallory")
	if VerifySubmission(mallory.PublicKey, s) {
		t.Fatal("submission verified under wrong key")
	}
}

func TestRegistrationSelfAuthenticating(t *testing.T) {
	id, _ := identity.New("bob")
	r := Registration{
		Version: APIVersion, UserID: id.UserID, Name: id.Name,
		PublicKey: "", Nonce: "n1", Timestamp: time.Now().Unix(),
	}
	SignRegistration(&r, id.Sign)
	if !VerifyRegistration(id.PublicKey, r) {
		t.Fatal("valid registration failed verification")
	}
	// The server independently derives the user_id from the public key and
	// must get the same value — nobody can register a key under another id.
	if identity.UserIDFromPublicKey(id.PublicKey) != r.UserID {
		t.Fatal("user_id does not match public key")
	}
}

func TestReplayDerivesElapsedFromLogNotClaim(t *testing.T) {
	// 50 correct chars over 6s → WPM = (50/5)/(6/60) = 100.
	log := buildLog(50, 6*time.Second)
	m := Replay(log)
	if m.WPM < 99 || m.WPM > 101 {
		t.Fatalf("replayed WPM = %v, want ~100", m.WPM)
	}
}

func TestAcceptRejectsInflatedWPM(t *testing.T) {
	log := buildLog(50, 6*time.Second) // true WPM ~100
	honest := Submission{WPM: 100, Log: log}
	if _, ok := Accept(honest); !ok {
		t.Fatal("honest submission rejected")
	}
	// Cheater claims 500 WPM but the log only supports ~100.
	cheat := Submission{WPM: 500, Log: log}
	if _, ok := Accept(cheat); ok {
		t.Fatal("inflated WPM accepted — replay check failed")
	}
	// Cheater also can't shrink ElapsedMS to fake it: Accept ignores the
	// claimed elapsed entirely and uses the log timestamps.
	cheat2 := Submission{WPM: 100, ElapsedMS: 1, Log: log}
	if m, ok := Accept(cheat2); !ok || m.WPM > 101 {
		t.Fatalf("elapsed spoof leaked through: ok=%v wpm=%v", ok, m.WPM)
	}
}

func TestLogDigestStableAndDistinct(t *testing.T) {
	a := buildLog(10, time.Second)
	b := buildLog(10, time.Second)
	if LogDigest(a) != LogDigest(b) {
		t.Fatal("identical logs should share a digest")
	}
	c := buildLog(11, time.Second)
	if LogDigest(a) == LogDigest(c) {
		t.Fatal("different logs should differ in digest")
	}
}
