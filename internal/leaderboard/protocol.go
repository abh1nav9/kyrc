// Package leaderboard defines the wire protocol shared by the kyrc CLI
// (client) and the leaderboard API (server), plus the client that submits
// scores. Keeping the signed-payload definition in ONE place is a security
// property: the client and server sign/verify byte-for-byte the same thing,
// so there is no room for a canonicalization mismatch that could be abused.
//
// Trust model recap:
//   - Every submission is Ed25519-signed by the user's private key.
//   - The server holds only public keys; it verifies the signature and
//     confirms user_id == UserIDFromPublicKey(pubkey) on registration.
//   - The submission carries the raw keystroke log; the server REPLAYS it
//     through the same pure engine and rejects any claimed WPM that does not
//     match the replay. So a client cannot report a fabricated WPM.
//   - A per-submission nonce + timestamp block replay of a captured request.
package leaderboard

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/abh1nav9/kyrc/internal/engine"
)

// APIVersion is bumped if the signed payload shape changes.
const APIVersion = 1

// Registration is sent once to associate a user_id with its public key and
// display name. It is self-authenticating: the server recomputes the
// user_id from the public key and verifies the signature over the payload,
// so nobody can register a public key under someone else's user_id.
type Registration struct {
	Version   int    `json:"version"`
	UserID    string `json:"user_id"`
	Name      string `json:"name"`
	PublicKey string `json:"public_key"` // hex-encoded ed25519 public key
	Nonce     string `json:"nonce"`
	Timestamp int64  `json:"timestamp"` // unix seconds
	Signature string `json:"signature"` // hex; over SignBytes(reg)
}

// Submission is a signed score push. WPM (and friends) are what the client
// claims; the server MUST recompute them from Log and reject on mismatch.
type Submission struct {
	Version int    `json:"version"`
	UserID  string `json:"user_id"`
	Name    string `json:"name"`

	Mode string `json:"mode"`
	// Claimed metrics — advisory; server recomputes from Log.
	WPM         float64 `json:"wpm"`
	RawWPM      float64 `json:"raw_wpm"`
	Accuracy    float64 `json:"accuracy"`
	Consistency float64 `json:"consistency"`
	ElapsedMS   int64   `json:"elapsed_ms"`

	// Log is the append-only keystroke record the server replays.
	Log []engine.Keystroke `json:"log"`

	Nonce     string `json:"nonce"`
	Timestamp int64  `json:"timestamp"`
	Signature string `json:"signature"` // hex; over SignBytes(sub)
}

// LeaderboardEntry is a public row returned to the CLI and the website.
// It intentionally contains no keystroke logs or keys — just display data.
type LeaderboardEntry struct {
	Rank     int     `json:"rank"`
	Name     string  `json:"name"`
	UserID   string  `json:"user_id"`
	WPM      float64 `json:"wpm"`
	Accuracy float64 `json:"accuracy"`
	AchievedAt int64 `json:"achieved_at"` // unix seconds
}

// canonical returns the deterministic bytes that get signed. We marshal a
// copy with Signature cleared so the signature never signs over itself, and
// rely on Go's json encoder producing struct fields in declaration order for
// a stable encoding. Both client and server call this exact function.
func canonicalRegistration(r Registration) []byte {
	r.Signature = ""
	b, _ := json.Marshal(r)
	return b
}

func canonicalSubmission(s Submission) []byte {
	s.Signature = ""
	b, _ := json.Marshal(s)
	return b
}

// SignRegistration signs r in place using signFn (identity.Sign).
func SignRegistration(r *Registration, signFn func([]byte) []byte) {
	r.Signature = hex.EncodeToString(signFn(canonicalRegistration(*r)))
}

// SignSubmission signs s in place using signFn (identity.Sign).
func SignSubmission(s *Submission, signFn func([]byte) []byte) {
	s.Signature = hex.EncodeToString(signFn(canonicalSubmission(*s)))
}

// VerifyRegistration checks the signature on r against pub.
func VerifyRegistration(pub ed25519.PublicKey, r Registration) bool {
	sig, err := hex.DecodeString(r.Signature)
	if err != nil {
		return false
	}
	return len(pub) == ed25519.PublicKeySize && ed25519.Verify(pub, canonicalRegistration(r), sig)
}

// VerifySubmission checks the signature on s against pub.
func VerifySubmission(pub ed25519.PublicKey, s Submission) bool {
	sig, err := hex.DecodeString(s.Signature)
	if err != nil {
		return false
	}
	return len(pub) == ed25519.PublicKeySize && ed25519.Verify(pub, canonicalSubmission(s), sig)
}

// LogDigest is a stable hash of a keystroke log, used server-side to reject
// duplicate submissions of the exact same run (idempotency / anti-spam).
func LogDigest(log []engine.Keystroke) string {
	b, _ := json.Marshal(log)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
