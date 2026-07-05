# Deploying the kyrc leaderboard (Render + Neon)

Follow these in order. Steps you must do are numbered; commands are copy-paste.

---

## 0. Rotate the Neon password (do this first)

The database password was exposed, so rotate it before anything else:

1. Neon console → your project → **Roles** → `neondb_owner` → **Reset password**.
2. Copy the **new** connection string (looks like
   `postgresql://neondb_owner:NEWPASS@ep-...neon.tech/neondb?sslmode=require`).

You'll paste this new string into Render (step 3) — never into code or git.

---

## 1. Apply the database schema

Needs `psql` locally (`brew install libpq && brew link --force libpq`).

```sh
cd server
export DATABASE_URL='<your NEW neon string>'
psql "$DATABASE_URL" -f schema.sql
# verify:
psql "$DATABASE_URL" -c '\dt'          # should list users, scores
```

---

## 2. Push the code

Render deploys from GitHub, so the code must be on `master` first:

```sh
git add -A
git commit -m "feat: leaderboard server + deploy config"
git push origin master
```

---

## 3. Create the Render service

The repo already includes [`render.yaml`](../render.yaml) and
[`server/Dockerfile`](Dockerfile), so this is a blueprint deploy:

1. render.com → **New → Blueprint**.
2. Connect the `abh1nav9/kyrc` repo. Render finds `render.yaml`.
3. When prompted for the env vars marked `sync: false`, set:
   - **`DATABASE_URL`** = your new Neon string (from step 0).
   - **`CORS_ORIGIN`** = your website's origin, e.g. `https://kyrc.dev`
     (use `*` temporarily if you don't have the final domain yet).
4. Click **Apply**. Render builds the Docker image and deploys.

> Why Docker? The server module uses `replace => ../` to share code with the
> CLI, so it must build from the repo root. The Dockerfile handles that; a
> plain Go build from `server/` alone would fail.

When it's live, Render gives you a URL like
`https://kyrc-leaderboard.onrender.com`. **Note it.**

---

## 4. Verify the server

```sh
SRV=https://kyrc-leaderboard.onrender.com   # your Render URL

curl "$SRV/healthz"        # → {"status":"ok"}
curl "$SRV/leaderboard"    # → {"leaderboard":[]}
```

> Render's free tier sleeps after inactivity; the first request may take
> ~30s to wake it. That's fine for a hobby leaderboard.

---

## 5. Point the CLI and site at your server

Two spots reference the API base — both default to `https://api.kyrc.dev`:

- **CLI**: `internal/leaderboard/client.go` → `DefaultBaseURL`.
  Users can override with `KYRC_LEADERBOARD_URL`, but bake in your real URL so
  it works with zero config.
- **Site**: reads `VITE_LEADERBOARD_URL` at build time
  (`site/.env` or the host's env). Set it to your Render URL.

Ask Claude to update `DefaultBaseURL` and add `site/.env` with your URL, or:

```sh
# site:
echo 'VITE_LEADERBOARD_URL=https://kyrc-leaderboard.onrender.com' > site/.env
```

---

## 6. Release the CLI + redeploy the site

```sh
# CLI feature release (minor bump — this is a real feature):
git tag v0.2.0 && git push origin v0.2.0

# site: redeploys automatically on push if hosted on Vercel/Netlify,
# else: cd site && npm run build  → upload dist/
```

---

## Smoke test the whole loop

```sh
kyrc login "Test User"     # creates account, prints user_id + passkey
kyrc                       # take a test (type the words)
kyrc sync                  # push your best
kyrc leaderboard           # you should see yourself
curl "$SRV/leaderboard"    # and via the API
```

## Rollback / notes

- No secrets live in the repo — everything sensitive is a Render env var.
- To wipe the leaderboard: `psql "$DATABASE_URL" -c 'TRUNCATE scores, users;'`
- Logs: Render dashboard → your service → **Logs**.
