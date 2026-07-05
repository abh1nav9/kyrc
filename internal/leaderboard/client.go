package leaderboard

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/abh1nav9/kyrc/internal/identity"
	"github.com/abh1nav9/kyrc/internal/store"
)

// DefaultBaseURL is the hosted leaderboard API. Overridable via the
// KYRC_LEADERBOARD_URL env var (read by the caller) for self-hosting/testing.
const DefaultBaseURL = "https://kyrc-server.onrender.com"

// Client talks to the leaderboard API. All methods are best-effort and
// context-bounded: kyrc is offline-first, so a network failure here must
// never break the app — callers log/ignore the error and carry on.
type Client struct {
	BaseURL string
	HTTP    *http.Client
}

// NewClient returns a Client with a short timeout so a hung network never
// blocks the CLI for long.
func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		BaseURL: baseURL,
		HTTP:    &http.Client{Timeout: 8 * time.Second},
	}
}

// Register associates the identity's user_id + public key with the server.
// Idempotent; safe to call on every sync.
func (c *Client) Register(ctx context.Context, id *identity.Identity) error {
	reg := Registration{
		Version:   APIVersion,
		UserID:    id.UserID,
		Name:      id.Name,
		PublicKey: hex.EncodeToString(id.PublicKey),
		Nonce:     newNonce(),
		Timestamp: time.Now().Unix(),
	}
	SignRegistration(&reg, id.Sign)
	return c.post(ctx, "/register", reg)
}

// Submit signs and pushes a single result. The server replays the log, so a
// tampered WPM is rejected server-side.
func (c *Client) Submit(ctx context.Context, id *identity.Identity, r store.Result) error {
	sub := Submission{
		Version:     APIVersion,
		UserID:      id.UserID,
		Name:        id.Name,
		Mode:        r.Mode,
		WPM:         r.WPM,
		RawWPM:      r.RawWPM,
		Accuracy:    r.Accuracy,
		Consistency: r.Consistency,
		ElapsedMS:   r.ElapsedMS,
		Log:         r.Log,
		Nonce:       newNonce(),
		Timestamp:   time.Now().Unix(),
	}
	SignSubmission(&sub, id.Sign)
	return c.post(ctx, "/submit", sub)
}

// SyncBest registers (idempotently) then submits the user's best result.
// This is the one call the CLI makes after a test when online. Returns nil
// on success; any error is safe to ignore (offline-first).
func (c *Client) SyncBest(ctx context.Context, id *identity.Identity, h *store.History) error {
	best, ok := h.Best()
	if !ok {
		return nil // nothing to push yet
	}
	if err := c.Register(ctx, id); err != nil {
		return fmt.Errorf("register: %w", err)
	}
	if err := c.Submit(ctx, id, best); err != nil {
		return fmt.Errorf("submit: %w", err)
	}
	return nil
}

// Top fetches the public leaderboard (best WPM per user), up to limit rows.
func (c *Client) Top(ctx context.Context, limit int) ([]LeaderboardEntry, error) {
	url := fmt.Sprintf("%s/leaderboard?limit=%d", c.BaseURL, limit)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("leaderboard: status %d", resp.StatusCode)
	}
	var out struct {
		Leaderboard []LeaderboardEntry `json:"leaderboard"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Leaderboard, nil
}

func (c *Client) post(ctx context.Context, path string, body any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var e struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&e)
		return fmt.Errorf("%s: status %d: %s", path, resp.StatusCode, e.Error)
	}
	return nil
}

// newNonce returns 16 random bytes hex-encoded. Combined with the timestamp
// and the server's skew window, it makes captured requests non-replayable.
func newNonce() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
