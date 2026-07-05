package identity

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// On-disk layout under the OS config dir (e.g. ~/.config/kyrc/):
//
//	identity.json   public: name, user_id, public key (safe to read/share)
//	key             private: the 32-byte seed, hex, file mode 0600
//	recovery.txt    human-readable card with user_id + recovery phrase
//
// The private key lives in its OWN 0600 file, separate from the public
// identity.json, so we can lock down its permissions precisely and never
// risk leaking it through a file meant to be readable.

const (
	identityFile = "identity.json"
	keyFile      = "key"
	recoveryFile = "recovery.txt"
)

// publicIdentity is the JSON shape of identity.json — no private material.
type publicIdentity struct {
	Name      string `json:"name"`
	UserID    string `json:"user_id"`
	PublicKey string `json:"public_key"` // hex
}

// Paths returns the three identity file paths under dir.
func Paths(dir string) (idPath, keyPath, recoveryPath string) {
	return filepath.Join(dir, identityFile),
		filepath.Join(dir, keyFile),
		filepath.Join(dir, recoveryFile)
}

// Exists reports whether an identity is already saved under dir.
func Exists(dir string) bool {
	_, keyPath, _ := Paths(dir)
	_, err := os.Stat(keyPath)
	return err == nil
}

// Save writes identity.json (0644), key (0600), and recovery.txt (0600) to
// dir, creating dir if needed. recovery.txt is the card we point users to.
func (id *Identity) Save(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	idPath, keyPath, recoveryPath := Paths(dir)

	pub := publicIdentity{
		Name:      id.Name,
		UserID:    id.UserID,
		PublicKey: hex.EncodeToString(id.PublicKey),
	}
	pb, err := json.MarshalIndent(pub, "", "  ")
	if err != nil {
		return err
	}
	if err := writeFileAtomic(idPath, pb, 0o644); err != nil {
		return err
	}

	// Private seed, hex, 0600.
	seedHex := hex.EncodeToString(id.privateKey.Seed())
	if err := writeFileAtomic(keyPath, []byte(seedHex+"\n"), 0o600); err != nil {
		return err
	}

	// Recovery card — the guided reference the docs tell users to open.
	if err := writeFileAtomic(recoveryPath, []byte(id.recoveryCard()), 0o600); err != nil {
		return err
	}
	return nil
}

// recoveryCard is the exact text written to recovery.txt.
func (id *Identity) recoveryCard() string {
	var b strings.Builder
	b.WriteString("kyrc account recovery\n")
	b.WriteString("=====================\n\n")
	b.WriteString("Keep this file private. Anyone with the recovery phrase below\n")
	b.WriteString("can log in as you. kyrc never sends it anywhere.\n\n")
	fmt.Fprintf(&b, "name:      %s\n", id.Name)
	fmt.Fprintf(&b, "user_id:   %s\n", id.UserID)
	b.WriteString("\nrecovery phrase (your \"passkey\"):\n\n")
	fmt.Fprintf(&b, "  %s\n\n", id.EncodePasskey())
	b.WriteString("To log in on another machine, run kyrc and choose \"restore\",\n")
	b.WriteString("then enter the user_id and recovery phrase above.\n")
	return b.String()
}

// LoadPublic reads just the public identity.json (name + user_id + pubkey),
// without the private key. Useful for display and for the server-side view.
func LoadPublic(dir string) (name, userID string, pub ed25519.PublicKey, err error) {
	idPath, _, _ := Paths(dir)
	b, err := os.ReadFile(idPath)
	if err != nil {
		return "", "", nil, err
	}
	var p publicIdentity
	if err := json.Unmarshal(b, &p); err != nil {
		return "", "", nil, err
	}
	pk, err := hex.DecodeString(p.PublicKey)
	if err != nil {
		return "", "", nil, fmt.Errorf("bad public key in %s: %w", idPath, err)
	}
	return p.Name, p.UserID, ed25519.PublicKey(pk), nil
}

// Load reads the full identity (including the private key) from dir. It also
// verifies the stored user_id/public key actually match the private key, so
// a tampered or mismatched identity file is rejected rather than trusted.
func Load(dir string) (*Identity, error) {
	_, keyPath, _ := Paths(dir)

	name, userID, pub, err := LoadPublic(dir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, err // caller distinguishes "no identity yet"
		}
		return nil, err
	}

	seedHex, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	seed, err := hex.DecodeString(strings.TrimSpace(string(seedHex)))
	if err != nil {
		return nil, fmt.Errorf("bad key file: %w", err)
	}
	id, err := fromSeed(name, seed)
	if err != nil {
		return nil, err
	}
	// Integrity: the derived identity must match the public file.
	if id.UserID != userID || !id.PublicKey.Equal(pub) {
		return nil, errors.New("identity files are inconsistent (key does not match identity.json)")
	}
	return id, nil
}

// writeFileAtomic writes via a temp file + rename with the given mode.
func writeFileAtomic(path string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".kyrc-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(mode); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
