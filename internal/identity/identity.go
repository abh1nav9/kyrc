// Package identity manages a kyrc user's cryptographic identity.
//
// Security model (why this is not spoofable):
//
//   - Each user has an Ed25519 keypair. The PRIVATE key never leaves the
//     device and is never sent to the server. The server only ever stores
//     the PUBLIC key.
//   - user_id is derived from the public key, so it is self-authenticating:
//     given a user_id and a signature, the server can verify the signer
//     holds the matching private key. Knowing someone's user_id (it's
//     public, shown on the leaderboard) grants no power to act as them.
//   - The "passkey" the user is told to save is a recovery phrase: a
//     mnemonic encoding of the 32-byte private seed. It is the ONLY way to
//     restore the account on another machine. Anyone with it controls the
//     account — hence it is shown once and stored locally in a 0600 file.
//
// Because authentication is a signature challenge (not a transmitted
// secret), there is no shared secret to steal in transit or from the
// server, and submissions cannot be forged or replayed for another user.
package identity

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"
)

// Identity is a user's full identity, including the private key. Treat the
// PrivateKey as secret; never marshal it into anything that leaves the box.
type Identity struct {
	Name       string
	UserID     string
	PublicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey
}

// seedLen is the Ed25519 private seed size.
const seedLen = ed25519.SeedSize // 32

// New creates a fresh identity for the given display name with a randomly
// generated keypair. The randomness source is crypto/rand, so the key (and
// therefore the derived user_id and recovery phrase) is unique with
// overwhelming probability every time.
func New(name string) (*Identity, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("name must not be empty")
	}
	seed := make([]byte, seedLen)
	if _, err := rand.Read(seed); err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}
	return fromSeed(name, seed)
}

// fromSeed builds an Identity deterministically from a 32-byte seed. Used by
// both New (random seed) and Restore (seed decoded from a recovery phrase).
func fromSeed(name string, seed []byte) (*Identity, error) {
	if len(seed) != seedLen {
		return nil, fmt.Errorf("seed must be %d bytes, got %d", seedLen, len(seed))
	}
	priv := ed25519.NewKeyFromSeed(seed)
	pub := priv.Public().(ed25519.PublicKey)
	return &Identity{
		Name:       name,
		UserID:     deriveUserID(pub),
		PublicKey:  pub,
		privateKey: priv,
	}, nil
}

// b32 is lowercase, no-padding base32 — compact and unambiguous for humans
// (no 0/O or 1/l confusion in the RFC4648 alphabet we downcase).
var b32 = base32.StdEncoding.WithPadding(base32.NoPadding)

// deriveUserID produces a stable, public identifier from the public key:
// the first 15 bytes of SHA-256(pubkey), base32-encoded and grouped as
// "kyrc-XXXXXX-XXXXXX-XXXXXX". It's a fingerprint of the key, so two users
// cannot share a user_id without sharing a key.
func deriveUserID(pub ed25519.PublicKey) string {
	sum := sha256.Sum256(pub)
	enc := strings.ToLower(b32.EncodeToString(sum[:15])) // 15 bytes -> 24 chars
	return "kyrc-" + enc[0:6] + "-" + enc[6:12] + "-" + enc[12:18] + "-" + enc[18:24]
}

// Sign returns an Ed25519 signature over msg using the private key. The
// server verifies this against the public key it has on file for UserID.
func (id *Identity) Sign(msg []byte) []byte {
	return ed25519.Sign(id.privateKey, msg)
}

// Verify checks a signature against a public key. Exported so the server
// (which imports this package) uses the exact same verification the client
// signs with — no room for a subtle mismatch.
func Verify(pub ed25519.PublicKey, msg, sig []byte) bool {
	return len(pub) == ed25519.PublicKeySize && ed25519.Verify(pub, msg, sig)
}

// UserIDFromPublicKey lets the server recompute the user_id from a stored
// public key and confirm it matches what the client claims.
func UserIDFromPublicKey(pub ed25519.PublicKey) string {
	return deriveUserID(pub)
}
