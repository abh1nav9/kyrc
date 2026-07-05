package identity

import (
	"strings"
	"testing"
)

func TestNewProducesUniqueIdentities(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 200; i++ {
		id, err := New("alice")
		if err != nil {
			t.Fatal(err)
		}
		if seen[id.UserID] {
			t.Fatalf("duplicate user_id generated: %s", id.UserID)
		}
		seen[id.UserID] = true
		if !strings.HasPrefix(id.UserID, "kyrc-") {
			t.Fatalf("user_id missing prefix: %s", id.UserID)
		}
	}
}

func TestNewRejectsEmptyName(t *testing.T) {
	if _, err := New("   "); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestSignVerifyRoundTrip(t *testing.T) {
	id, _ := New("bob")
	msg := []byte("wpm=99;ts=123")
	sig := id.Sign(msg)
	if !Verify(id.PublicKey, msg, sig) {
		t.Fatal("valid signature failed to verify")
	}
	// Tampered message must fail.
	if Verify(id.PublicKey, []byte("wpm=100;ts=123"), sig) {
		t.Fatal("signature verified against tampered message")
	}
	// Another user's key must not verify — the core anti-spoof property.
	other, _ := New("mallory")
	if Verify(other.PublicKey, msg, sig) {
		t.Fatal("signature verified under wrong public key")
	}
}

func TestRecoveryPhraseRoundTrip(t *testing.T) {
	id, _ := New("carol")
	phrase := id.EncodePasskey()

	restored, err := Restore("carol", phrase)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if restored.UserID != id.UserID {
		t.Fatalf("restored user_id %s != original %s", restored.UserID, id.UserID)
	}
	if !restored.PublicKey.Equal(id.PublicKey) {
		t.Fatal("restored public key differs")
	}
	// A restored identity must be able to produce signatures the original's
	// public key verifies (proves the private key is identical).
	msg := []byte("hello")
	if !Verify(id.PublicKey, msg, restored.Sign(msg)) {
		t.Fatal("restored key cannot sign for original identity")
	}
}

func TestRecoveryPhraseIsFormatted(t *testing.T) {
	id, _ := New("dave")
	phrase := id.EncodePasskey()
	if !strings.Contains(phrase, "-") {
		t.Fatalf("phrase should be dash-grouped: %s", phrase)
	}
}

func TestRestoreToleratesFormattingAndCatchesTypos(t *testing.T) {
	id, _ := New("erin")
	phrase := id.EncodePasskey()

	// Formatting tolerance: lowercase, spaces, prefix should all restore.
	messy := "kyrc-recovery: " + strings.ToLower(strings.ReplaceAll(phrase, "-", " "))
	if r, err := Restore("erin", messy); err != nil || r.UserID != id.UserID {
		t.Fatalf("messy-but-valid phrase failed: err=%v", err)
	}

	// Typo detection: flip a character; checksum should reject.
	runes := []rune(strings.ReplaceAll(phrase, "-", ""))
	if runes[0] == 'A' {
		runes[0] = 'B'
	} else {
		runes[0] = 'A'
	}
	if _, err := Restore("erin", string(runes)); err == nil {
		t.Fatal("expected checksum failure on typo")
	}
}

func TestSaveLoadRoundTripAndIntegrity(t *testing.T) {
	dir := t.TempDir()
	id, _ := New("frank")
	if err := id.Save(dir); err != nil {
		t.Fatalf("save: %v", err)
	}
	if !Exists(dir) {
		t.Fatal("Exists false after Save")
	}

	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.UserID != id.UserID || !loaded.PublicKey.Equal(id.PublicKey) {
		t.Fatal("loaded identity mismatch")
	}

	// The private key file must be 0600.
	_, keyPath, recoveryPath := Paths(dir)
	fi, err := statMode(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	if fi&0o077 != 0 {
		t.Fatalf("key file is group/other-accessible: mode %o", fi)
	}
	// Recovery card should exist and contain the user_id.
	card, err := readFile(recoveryPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(card, id.UserID) {
		t.Fatal("recovery card missing user_id")
	}
}

func TestUserIDFromPublicKeyMatches(t *testing.T) {
	id, _ := New("grace")
	if UserIDFromPublicKey(id.PublicKey) != id.UserID {
		t.Fatal("server-side user_id derivation disagrees with client")
	}
}
