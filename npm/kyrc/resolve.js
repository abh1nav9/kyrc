// Resolve the path to the platform-specific kyrc binary.
//
// Strategy (same as esbuild): each OS/arch ships as its own optional
// dependency package (@kyrc/<os>-<arch>) whose package.json declares the
// matching `os`/`cpu`, so npm installs ONLY the one that fits the host and
// silently skips the rest. Here we locate whichever one landed.
const path = require("path");

// Map Node's platform/arch naming to our package suffixes.
const platformKey = `${process.platform}-${process.arch}`;

// The binary is named kyrc (kyrc.exe on Windows) inside the sub-package.
const binName = process.platform === "win32" ? "kyrc.exe" : "kyrc";

function resolveBinaryPath() {
  const pkg = `@kyrc/${platformKey}`;
  try {
    // require.resolve finds the sub-package's own package.json, giving us
    // its install location regardless of hoisting.
    const pkgJson = require.resolve(`${pkg}/package.json`);
    return path.join(path.dirname(pkgJson), "bin", binName);
  } catch (e) {
    return null;
  }
}

module.exports = { resolveBinaryPath, platformKey, binName };
