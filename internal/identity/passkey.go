package identity

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"
)

// The recovery phrase ("passkey") encodes the 32-byte private seed as a
// checksummed, grouped base32 string that a human can copy from a file or
// type back in. Format:
//
//	kyrc-recovery: AAAA-BBBB-CCCC-... (mixed 4-char groups)
//
// We use base32 (case-insensitive, no ambiguous padding) rather than a
// word mnemonic to stay dependency-free and avoid shipping a wordlist,
// while a trailing checksum byte catches transcription errors before we
// ever try to build a key from a mistyped phrase.

// EncodePasskey turns the identity's private seed into a recovery phrase.
func (id *Identity) EncodePasskey() string {
	seed := id.privateKey.Seed() // 32 bytes
	return encodePasskey(seed)
}

func encodePasskey(seed []byte) string {
	// Append a 1-byte checksum = first byte of SHA-256(seed).
	sum := sha256.Sum256(seed)
	payload := make([]byte, 0, len(seed)+1)
	payload = append(payload, seed...)
	payload = append(payload, sum[0])

	raw := strings.ToUpper(b32.EncodeToString(payload)) // no padding
	// Group into 4-char blocks for readability.
	var b strings.Builder
	for i, r := range raw {
		if i > 0 && i%4 == 0 {
			b.WriteByte('-')
		}
		b.WriteRune(r)
	}
	return b.String()
}

// Restore rebuilds an Identity from a display name and a recovery phrase.
// It validates the checksum, so a mistyped phrase fails loudly instead of
// silently producing a different (wrong) account.
func Restore(name, phrase string) (*Identity, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("name must not be empty")
	}
	seed, err := decodePasskey(phrase)
	if err != nil {
		return nil, err
	}
	return fromSeed(name, seed)
}

func decodePasskey(phrase string) ([]byte, error) {
	// Normalize: strip any prefix, spaces, dashes; uppercase.
	p := strings.TrimSpace(phrase)
	if i := strings.IndexByte(p, ':'); i >= 0 {
		p = p[i+1:] // drop a "kyrc-recovery:" style prefix
	}
	p = strings.NewReplacer("-", "", " ", "", "\t", "", "\n", "").Replace(p)
	p = strings.ToUpper(p)
	if p == "" {
		return nil, errors.New("empty recovery phrase")
	}

	payload, err := b32.DecodeString(p)
	if err != nil {
		return nil, fmt.Errorf("recovery phrase is not valid (check for typos): %w", err)
	}
	if len(payload) != seedLen+1 {
		return nil, fmt.Errorf("recovery phrase has wrong length (expected %d bytes, got %d) — likely a typo", seedLen+1, len(payload))
	}
	seed := payload[:seedLen]
	want := payload[seedLen]
	sum := sha256.Sum256(seed)
	if sum[0] != want {
		return nil, errors.New("recovery phrase checksum failed — it was mistyped")
	}
	return seed, nil
}
