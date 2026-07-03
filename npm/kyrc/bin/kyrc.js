#!/usr/bin/env node
// Launcher: exec the native kyrc binary, forwarding args, stdio, and the
// exit code. We use a real process (not a wrapper that reads stdout) so the
// binary owns the TTY directly — essential for a raw-mode terminal app.
"use strict";

const { spawnSync } = require("child_process");
const { resolveBinaryPath, platformKey } = require("../resolve.js");

const bin = resolveBinaryPath();
if (!bin) {
  console.error(
    `kyrc: no prebuilt binary for your platform (${platformKey}).\n` +
      `This usually means the optional dependency failed to install.\n` +
      `Try: npm install kyrc --force  (or grab a binary from GitHub releases).`
  );
  process.exit(1);
}

// stdio:'inherit' hands the TTY straight to kyrc so raw-mode input and the
// alternate screen work exactly as if run directly.
const result = spawnSync(bin, process.argv.slice(2), { stdio: "inherit" });

if (result.error) {
  console.error(`kyrc: failed to launch binary: ${result.error.message}`);
  process.exit(1);
}
process.exit(result.status === null ? 1 : result.status);
