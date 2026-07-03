import { useState } from "react";
import { motion } from "motion/react";

// InstallTabs shows the install one-liner across JS package managers and
// offers copy-to-clipboard. Homebrew is intentionally absent — kyrc ships
// via the npm ecosystem (a Go static binary delivered through per-platform
// optional-dependency packages).
const MANAGERS: { id: string; label: string; cmd: string }[] = [
  { id: "npm", label: "npm", cmd: "npm i -g kyrc" },
  { id: "bun", label: "bun", cmd: "bun add -g kyrc" },
  { id: "pnpm", label: "pnpm", cmd: "pnpm add -g kyrc" },
  { id: "npx", label: "npx", cmd: "npx kyrc" },
];

export function InstallTabs() {
  const [active, setActive] = useState(MANAGERS[0].id);
  const [copied, setCopied] = useState(false);
  const current = MANAGERS.find((m) => m.id === active)!;

  async function copy() {
    try {
      await navigator.clipboard.writeText(current.cmd);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1400);
    } catch {
      /* clipboard blocked; no-op */
    }
  }

  return (
    <div className="max-w-md overflow-hidden rounded-[10px] border border-border bg-bg-panel">
      <div
        className="flex border-b border-border bg-bg-soft"
        role="tablist"
        aria-label="Package manager"
      >
        {MANAGERS.map((m) => (
          <button
            key={m.id}
            role="tab"
            aria-selected={active === m.id}
            onClick={() => setActive(m.id)}
            className={`relative flex-1 cursor-pointer py-2.5 font-mono text-[13px] transition-colors ${
              active === m.id
                ? "text-accent"
                : "text-faint hover:text-dim"
            }`}
          >
            {m.label}
            {active === m.id && (
              <motion.span
                layoutId="install-underline"
                className="absolute inset-x-0 bottom-0 h-0.5 bg-accent"
                transition={{ type: "spring", stiffness: 500, damping: 35 }}
              />
            )}
          </button>
        ))}
      </div>
      <div className="flex items-center justify-between px-4 py-3.5 text-sm">
        <code className="font-mono">
          <span className="mr-1.5 text-term">$</span>
          {current.cmd}
        </code>
        <button
          onClick={copy}
          aria-label="Copy install command"
          className="cursor-pointer rounded-md border border-border bg-bg-soft px-3 py-1.5 font-mono text-xs text-dim transition-colors hover:border-accent-soft hover:text-accent"
        >
          {copied ? "copied ✓" : "copy"}
        </button>
      </div>
    </div>
  );
}
