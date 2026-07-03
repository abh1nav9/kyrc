#!/usr/bin/env node
// Stamp the meta-package (@kyrc/kyrc) version AND re-pin every platform
// optionalDependency to the same version, from KYRC_VERSION. This keeps a
// single source of truth (the git tag) — the meta-package and all platform
// packages always ship in lockstep, so an install can never resolve a
// version-mismatched binary.
"use strict";

const fs = require("fs");
const path = require("path");

const version = process.env.KYRC_VERSION;
if (!version) {
  console.error("KYRC_VERSION is required (e.g. 0.1.2)");
  process.exit(1);
}
// Accept a leading "v" from a git tag and strip it.
const clean = version.replace(/^v/, "");

const metaPath = path.join(__dirname, "..", "kyrc", "package.json");
const meta = JSON.parse(fs.readFileSync(metaPath, "utf8"));

meta.version = clean;
for (const dep of Object.keys(meta.optionalDependencies || {})) {
  meta.optionalDependencies[dep] = clean;
}

fs.writeFileSync(metaPath, JSON.stringify(meta, null, 2) + "\n");
console.log(`meta @kyrc/kyrc -> ${clean}, optionalDependencies pinned to ${clean}`);
