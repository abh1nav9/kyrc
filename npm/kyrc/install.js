// Postinstall sanity check. The actual binary arrives via the platform
// optional-dependency package, so there's nothing to DOWNLOAD here — this
// just fails loudly (but non-fatally) if npm skipped every platform package,
// e.g. behind a restrictive proxy or with --no-optional. A clear message
// now beats a cryptic "binary not found" at first run.
"use strict";

const { resolveBinaryPath, platformKey } = require("./resolve.js");
const fs = require("fs");

const bin = resolveBinaryPath();
if (!bin || !fs.existsSync(bin)) {
  console.error(
    `\n[kyrc] Prebuilt binary for ${platformKey} was not installed.\n` +
      `       If your setup blocks optional dependencies, reinstall with\n` +
      `       optional deps enabled, or grab a binary from:\n` +
      `       https://github.com/abh1nav9/kyrc/releases\n`
  );
  // Non-fatal: don't abort the whole npm install over this.
  process.exit(0);
}

// Ensure the binary is executable (npm can strip the bit on some setups).
try {
  fs.chmodSync(bin, 0o755);
} catch (_) {
  /* best effort */
}
