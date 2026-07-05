# kyrc leaderboard API

The service in front of Neon/Postgres for the kyrc leaderboard. It is the
**only** component that holds the database URL — the CLI talks to this API,
never to Postgres directly (shipping DB creds in a public binary would let
anyone bypass every check).

## What it enforces

- **Account authenticity** — every request is Ed25519-signed by the user's
  private key; the server verifies against the public key on file and
  confirms `user_id == fingerprint(public_key)`. No shared secret is ever
  transmitted, so accounts cannot be impersonated.
- **Score integrity** — submissions carry the raw keystroke log; the server
  **replays** it (`leaderboard.Accept`) and stores the *replayed* metrics,
  rejecting any claimed WPM that doesn't match. Elapsed time is taken from
  the log's own timestamps, so it can't be spoofed.
- **Replay resistance** — a per-request nonce + timestamp (±10 min skew) and
  a per-run `log_digest` uniqueness constraint prevent re-submitting a
  captured request or double-counting a run.

## Endpoints

| Method | Path           | Body                     | Purpose |
|--------|----------------|--------------------------|---------|
| POST   | `/register`    | `Registration` (signed)  | Associate user_id ↔ public key (idempotent) |
| POST   | `/submit`      | `Submission` (signed)    | Push a score; server replays + stores |
| GET    | `/leaderboard` | `?limit=` (1–500)        | Public top-N by best WPM |
| GET    | `/healthz`     | —                        | Health check |

## Run locally

```sh
# 1. Apply the schema to your database (once):
psql "$DATABASE_URL" -f schema.sql

# 2. Set the connection string in the ENVIRONMENT (never commit it):
export DATABASE_URL='postgresql://USER:PASSWORD@HOST/db?sslmode=require'
export CORS_ORIGIN='https://your-kyrc-site'   # optional; default *

# 3. Run:
go run .
# → kyrc-leaderboard listening on :8080
```

## Deploy

Any host that runs a Go binary + reaches Neon works (Fly.io, Railway,
Render, a VPS). Provide `DATABASE_URL` (and optionally `PORT`, `CORS_ORIGIN`)
as environment variables / secrets — **never** hardcode them.

> ⚠️ If a Neon password was ever pasted into a chat, an email, or a commit,
> rotate it in the Neon console before deploying.
