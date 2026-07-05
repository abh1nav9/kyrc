# Distribution & release setup

`git tag vX.Y.Z && git push origin vX.Y.Z` fans kyrc out to every native
installer via `.github/workflows/release.yml` + `.goreleaser.yaml`. Most of it
is automatic, but each channel needs **one-time setup** first. This file is the
checklist. Nothing here is done automatically — the release will simply skip
(or fail) a channel whose secret/repo is missing.

## TL;DR — what a user runs once everything is set up

```sh
brew install abh1nav9/tap/kyrc                                    # macOS / Linux
scoop bucket add kyrc https://github.com/abh1nav9/scoop-bucket    # Windows
  && scoop install kyrc
winget install abh1nav9.kyrc                                      # Windows
yay -S kyrc-bin                                                   # Arch
snap install kyrc                                                 # Linux
npm i -g @kyrc/kyrc                                               # anywhere w/ Node
# apt / dnf: see the landing page at https://abh1nav9.github.io/kyrc/
```

---

## One-time setup checklist

Do these before the first tag that should hit all channels. Each row is
independent; skip a channel and the rest still work.

### 1. Repos to create (empty is fine — CI populates them)

| Repo | Purpose | Notes |
|------|---------|-------|
| `abh1nav9/homebrew-tap`  | Homebrew cask lives here | Public. `brew install abh1nav9/tap/kyrc` resolves `homebrew-<X>` → tap `X`. |
| `abh1nav9/scoop-bucket`  | Scoop manifest lives here | Public. |
| `abh1nav9/winget-pkgs`   | **Fork** of `microsoft/winget-pkgs` | CI pushes a branch here, then opens a PR upstream. |

### 2. GitHub Pages (for apt / dnf)

- Repo **Settings → Pages → Source = `gh-pages` branch**. The first release
  creates the branch; enable Pages after that run (or pre-create an empty
  `gh-pages` branch so the very first release publishes immediately).

### 3. External accounts

| Account | For | What you need |
|---------|-----|---------------|
| ~~**AUR**~~ (aur.archlinux.org) | ~~`yay -S kyrc-bin`~~ | **Disabled** — AUR registration is shut down during the Arch "Atomic Arch" supply-chain cleanup. The `aurs:` block in `.goreleaser.yaml` is commented out; re-enable + set `AUR_KEY` once registration reopens. |
| **Snap Store** (snapcraft.io) | `snap install kyrc` | Register the name `kyrc` (`snapcraft register kyrc`), then `snapcraft export-login` → `SNAPCRAFT_STORE_CREDENTIALS`. |

### 4. GPG key (signs the apt/rpm repositories)

```sh
gpg --full-generate-key            # RSA 4096, no expiry is fine for this
gpg --armor --export-secret-keys <KEYID>   # → GPG_PRIVATE_KEY secret
# GPG_PASSPHRASE = the key's passphrase (set to "" if you made it passphrase-less)
```

### 5. Repository secrets

`Settings → Secrets and variables → Actions → New repository secret`:

| Secret | Value | Used by |
|--------|-------|---------|
| `NPM_TOKEN` | npm **Automation** token (bypasses 2FA), scope `@kyrc` | npm publish |
| `HOMEBREW_TAP_TOKEN` | PAT with `contents:write` on `abh1nav9/homebrew-tap` | Homebrew |
| `SCOOP_BUCKET_TOKEN` | PAT with `contents:write` on `abh1nav9/scoop-bucket` | Scoop |
| `WINGET_TOKEN` | PAT with `contents:write` on your `winget-pkgs` fork | WinGet |
| `AUR_KEY` | SSH **private** key registered with your AUR account | AUR |
| `SNAPCRAFT_STORE_CREDENTIALS` | output of `snapcraft export-login` | Snap |
| `GPG_PRIVATE_KEY` | ASCII-armored private key (see §4) | apt/rpm signing |
| `GPG_PASSPHRASE` | passphrase for that key (or empty) | apt/rpm signing |

> A dedicated token is required for the tap/bucket/winget repos because the
> default `GITHUB_TOKEN` can only write to *this* repo, not to a different one.

---

## What each channel needs, in one glance

| Channel | Auto from tag? | Blocking prerequisite |
|---------|----------------|-----------------------|
| GitHub Release (archives + `.deb`/`.rpm` files) | ✅ | none (uses built-in `GITHUB_TOKEN`) |
| npm | ✅ | `NPM_TOKEN` |
| Homebrew | ✅ | `homebrew-tap` repo + `HOMEBREW_TAP_TOKEN` |
| Scoop | ✅ | `scoop-bucket` repo + `SCOOP_BUCKET_TOKEN` |
| WinGet | ✅ (opens a PR) | `winget-pkgs` fork + `WINGET_TOKEN`; **first submit is manually reviewed by Microsoft** |
| AUR | ✅ | AUR account + `AUR_KEY` |
| Snap | ✅ | registered name + `SNAPCRAFT_STORE_CREDENTIALS` |
| apt / dnf (Pages repo) | ✅ | GPG secrets + Pages enabled |

## Verifying a release

After CI is green:

```sh
# Homebrew (from a mac/Linux box):
brew install abh1nav9/tap/kyrc && kyrc --version

# apt (Debian/Ubuntu container):
# follow https://abh1nav9.github.io/kyrc/  then: kyrc --version

# raw package, no repo:
curl -LO https://github.com/abh1nav9/kyrc/releases/latest/download/kyrc_<ver>_linux_amd64.deb
sudo dpkg -i kyrc_*.deb && kyrc --version
```

## Notes / gotchas

- **WinGet is not instant.** GoReleaser opens a PR to `microsoft/winget-pkgs`;
  a human reviews the first version of a new package. Later versions usually
  auto-merge. So `winget install` starts working only after that PR merges.
- **macOS Gatekeeper.** The binary is unsigned. The Homebrew cask strips the
  quarantine xattr on install (`post.install` hook), so `kyrc` runs without a
  prompt. Users who download the raw tarball may need
  `xattr -dr com.apple.quarantine ./kyrc` once.
- **apt/rpm repos accumulate.** The pages job uses `keep_files: true` +
  `reprepro`/`createrepo_c --update`, so older versions stay installable.
- **Snap review.** `strict` confinement with only `home` + `network` plugs
  should pass automated review; if the store flags it, loosen or explain in
  the store listing.
```
