-- kyrc leaderboard schema (Neon / Postgres).
--
-- Apply once against your database:
--   psql "$DATABASE_URL" -f schema.sql
--
-- Design notes:
--   * users holds the PUBLIC key only. No secret is ever stored here, so a
--     database leak cannot be used to impersonate anyone — signatures are
--     verified against these public keys, and the private keys never leave
--     users' devices.
--   * scores keeps every accepted submission (server-REPLAYED metrics, not
--     client-claimed). The leaderboard view surfaces each user's best.
--   * log_digest gives us idempotency: the same run can't be double-counted.

CREATE TABLE IF NOT EXISTS users (
    user_id     TEXT PRIMARY KEY,           -- derived from public_key
    name        TEXT NOT NULL,
    public_key  TEXT NOT NULL UNIQUE,       -- hex ed25519 public key
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS scores (
    id           BIGSERIAL PRIMARY KEY,
    user_id      TEXT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    wpm          DOUBLE PRECISION NOT NULL, -- server-replayed, authoritative
    raw_wpm      DOUBLE PRECISION NOT NULL,
    accuracy     DOUBLE PRECISION NOT NULL,
    consistency  DOUBLE PRECISION NOT NULL,
    mode         TEXT NOT NULL,
    log_digest   TEXT NOT NULL,             -- sha256 of the keystroke log
    achieved_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, log_digest)            -- same run submitted twice = no-op
);

CREATE INDEX IF NOT EXISTS scores_wpm_idx ON scores (wpm DESC);
CREATE INDEX IF NOT EXISTS scores_user_idx ON scores (user_id);

-- One row per user: their best accepted WPM. This is what the CLI and the
-- website read for the public leaderboard.
CREATE OR REPLACE VIEW leaderboard AS
SELECT DISTINCT ON (s.user_id)
       u.name,
       s.user_id,
       s.wpm,
       s.accuracy,
       s.achieved_at
FROM scores s
JOIN users u USING (user_id)
ORDER BY s.user_id, s.wpm DESC;
