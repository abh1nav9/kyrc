package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/abh1nav9/kyrc/internal/identity"
	"github.com/abh1nav9/kyrc/internal/leaderboard"
)

type server struct {
	db   *pgxpool.Pool
	cors string
}

// maxClockSkew bounds how stale/future a submission timestamp may be. It
// makes captured requests useless to replay after a few minutes.
const maxClockSkew = 10 * time.Minute

func (s *server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleRegister associates a user_id with its public key. Idempotent: a
// repeat registration of the same key/name succeeds without error so a user
// restoring on a new machine just works.
func (s *server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "POST only")
		return
	}
	var reg leaderboard.Registration
	if err := json.NewDecoder(r.Body).Decode(&reg); err != nil {
		writeErr(w, http.StatusBadRequest, "bad json")
		return
	}
	if reg.Version != leaderboard.APIVersion {
		writeErr(w, http.StatusBadRequest, "unsupported api version")
		return
	}
	if !withinSkew(reg.Timestamp) {
		writeErr(w, http.StatusBadRequest, "timestamp out of range")
		return
	}
	pub, err := hex.DecodeString(reg.PublicKey)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "bad public key")
		return
	}
	// The user_id MUST be the fingerprint of this public key. This is what
	// stops anyone registering a key under someone else's id.
	if identity.UserIDFromPublicKey(pub) != reg.UserID {
		writeErr(w, http.StatusBadRequest, "user_id does not match public key")
		return
	}
	if !leaderboard.VerifyRegistration(pub, reg) {
		writeErr(w, http.StatusUnauthorized, "signature verification failed")
		return
	}

	_, err = s.db.Exec(r.Context(), `
		INSERT INTO users (user_id, name, public_key)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE SET name = EXCLUDED.name
		WHERE users.public_key = EXCLUDED.public_key
	`, reg.UserID, reg.Name, reg.PublicKey)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"user_id": reg.UserID})
}

// handleSubmit accepts a signed score, verifies it against the stored public
// key, REPLAYS the log to get authoritative metrics, and stores those.
func (s *server) handleSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "POST only")
		return
	}
	var sub leaderboard.Submission
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		writeErr(w, http.StatusBadRequest, "bad json")
		return
	}
	if sub.Version != leaderboard.APIVersion {
		writeErr(w, http.StatusBadRequest, "unsupported api version")
		return
	}
	if !withinSkew(sub.Timestamp) {
		writeErr(w, http.StatusBadRequest, "timestamp out of range")
		return
	}

	// Look up the registered public key for this user.
	pub, err := s.publicKey(r.Context(), sub.UserID)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, "unknown user_id — register first")
		return
	}
	if !leaderboard.VerifySubmission(pub, sub) {
		writeErr(w, http.StatusUnauthorized, "signature verification failed")
		return
	}

	// Anti-cheat: recompute metrics from the log; reject fabricated WPM.
	metrics, ok := leaderboard.Accept(sub)
	if !ok {
		writeErr(w, http.StatusBadRequest, "score does not match its keystroke log")
		return
	}

	digest := leaderboard.LogDigest(sub.Log)
	_, err = s.db.Exec(r.Context(), `
		INSERT INTO scores (user_id, wpm, raw_wpm, accuracy, consistency, mode, log_digest)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, log_digest) DO NOTHING
	`, sub.UserID, metrics.WPM, metrics.RawWPM, metrics.Accuracy, metrics.Consistency, sub.Mode, digest)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"accepted": true, "wpm": metrics.WPM})
}

// handleLeaderboard returns the top N users by best WPM. Public, read-only,
// no auth. Query param ?limit= (default 100, max 500).
func (s *server) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := parseLimit(v); err == nil {
			limit = n
		}
	}
	rows, err := s.db.Query(r.Context(), `
		SELECT name, user_id, wpm, accuracy, achieved_at
		FROM leaderboard
		ORDER BY wpm DESC
		LIMIT $1
	`, limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db error")
		return
	}
	defer rows.Close()

	entries := []leaderboard.LeaderboardEntry{}
	rank := 0
	for rows.Next() {
		rank++
		var e leaderboard.LeaderboardEntry
		var at time.Time
		if err := rows.Scan(&e.Name, &e.UserID, &e.WPM, &e.Accuracy, &at); err != nil {
			writeErr(w, http.StatusInternalServerError, "db scan error")
			return
		}
		e.Rank = rank
		e.AchievedAt = at.Unix()
		entries = append(entries, e)
	}
	writeJSON(w, http.StatusOK, map[string]any{"leaderboard": entries})
}

func (s *server) publicKey(ctx context.Context, userID string) ([]byte, error) {
	var hexKey string
	err := s.db.QueryRow(ctx, `SELECT public_key FROM users WHERE user_id = $1`, userID).Scan(&hexKey)
	if err != nil {
		return nil, err // pgx.ErrNoRows for unknown user
	}
	return hex.DecodeString(hexKey)
}

func withinSkew(ts int64) bool {
	d := time.Since(time.Unix(ts, 0))
	if d < 0 {
		d = -d
	}
	return d <= maxClockSkew
}
