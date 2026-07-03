#!/usr/bin/env node
// Generate the per-platform npm sub-packages (@kyrc/<os>-<arch>) from
// GoReleaser's dist/ output, then stage them for publish. Run AFTER
// `goreleaser release --snapshot` (or a real release) so the binaries exist.
//
// This mirrors esbuild's model: one thin `kyrc` meta-package + N tiny
// platform packages, each declaring `os`/`cpu` so npm installs exactly one.
"use strict";

const fs = require("fs");
const path = require("path");

const VERSION = process.env.KYRC_VERSION || require("../kyrc/package.json").version;
const DIST = process.env.KYRC_DIST || path.join(__dirname, "..", "..", "dist");
const OUT = path.join(__dirname, "..", "packages");

// (npm os, npm cpu) -> GoReleaser (goos, goarch) + binary name.
const TARGETS = [
  { os: "darwin", cpu: "arm64", goos: "darwin", goarch: "arm64", bin: "kyrc" },
  { os: "darwin", cpu: "x64", goos: "darwin", goarch: "amd64", bin: "kyrc" },
  { os: "linux", cpu: "arm64", goos: "linux", goarch: "arm64", bin: "kyrc" },
  { os: "linux", cpu: "x64", goos: "linux", goarch: "amd64", bin: "kyrc" },
  { os: "win32", cpu: "x64", goos: "windows", goarch: "amd64", bin: "kyrc.exe" },
];

// GoReleaser lays binaries out under dist/<build-id>_<goos>_<goarch>[_v1]/.
// Find the built binary for a target by scanning dist for the right folder.
function findBinary(t) {
  const entries = fs.readdirSync(DIST, { withFileTypes: true });
  for (const e of entries) {
    if (!e.isDirectory()) continue;
    if (e.name.includes(`_${t.goos}_${t.goarch}`)) {
      const p = path.join(DIST, e.name, t.bin);
      if (fs.existsSync(p)) return p;
    }
  }
  return null;
}

function writePackage(t) {
  const pkgDir = path.join(OUT, `${t.os}-${t.cpu}`);
  const binDir = path.join(pkgDir, "bin");
  fs.mkdirSync(binDir, { recursive: true });

  const src = findBinary(t);
  if (!src) {
    console.warn(`[skip] no binary for ${t.os}-${t.cpu} (looked in ${DIST})`);
    return false;
  }
  fs.copyFileSync(src, path.join(binDir, t.bin));
  fs.chmodSync(path.join(binDir, t.bin), 0o755);

  const pkg = {
    name: `@kyrc/${t.os}-${t.cpu}`,
    version: VERSION,
    description: `kyrc binary for ${t.os}-${t.cpu}`,
    // These two fields are the whole trick: npm installs this package ONLY
    // when the host matches, and silently skips it otherwise.
    os: [t.os],
    cpu: [t.cpu],
    license: "MIT",
    files: [`bin/${t.bin}`],
  };
  fs.writeFileSync(
    path.join(pkgDir, "package.json"),
    JSON.stringify(pkg, null, 2) + "\n"
  );
  console.log(`[ok] staged @kyrc/${t.os}-${t.cpu}`);
  return true;
}

fs.rmSync(OUT, { recursive: true, force: true });
fs.mkdirSync(OUT, { recursive: true });
let n = 0;
for (const t of TARGETS) if (writePackage(t)) n++;
console.log(`\nStaged ${n}/${TARGETS.length} platform packages in ${OUT}`);
console.log("Publish each with: npm publish --access public");
