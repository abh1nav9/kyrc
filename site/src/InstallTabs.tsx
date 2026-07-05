import { useState } from "react";
import { motion } from "motion/react";

// InstallTabs shows the install one-liner across every native package
// manager kyrc ships to, with copy-to-clipboard. A single `git tag` fans
// the Go static binary out to all of these (see docs/DISTRIBUTION.md).
// `note` renders a second line for managers that need a one-time setup step.
const MANAGERS: {
  id: string;
  label: string;
  cmd: string;
  note?: string;
}[] = [
  { id: "brew", label: "brew", cmd: "brew install abh1nav9/tap/kyrc" },
  {
    id: "scoop",
    label: "scoop",
    cmd: "scoop install kyrc",
    note: "scoop bucket add kyrc https://github.com/abh1nav9/scoop-bucket",
  },
  { id: "winget", label: "winget", cmd: "winget install abh1nav9.kyrc" },
  { id: "aur", label: "aur", cmd: "yay -S kyrc-bin" },
  { id: "snap", label: "snap", cmd: "snap install kyrc" },
  { id: "npm", label: "npm", cmd: "npm i -g @kyrc/kyrc" },
  { id: "bun", label: "bun", cmd: "bun add -g @kyrc/kyrc" },
  { id: "pnpm", label: "pnpm", cmd: "pnpm add -g @kyrc/kyrc" },
  { id: "npx", label: "npx", cmd: "npx @kyrc/kyrc" },
];

export function InstallTabs() {
  const [active, setActive] = useState(MANAGERS[0].id);
  const [copied, setCopied] = useState(false);
  const current = MANAGERS.find((m) => m.id === active)!;

  // Copy the setup note (if any) plus the install command, so pasting a
  // manager like scoop that needs `bucket add` first Just Works.
  async function copy() {
    try {
      const text = current.note
        ? `${current.note}\n${current.cmd}`
        : current.cmd;
      await navigator.clipboard.writeText(text);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1400);
    } catch {
      /* clipboard blocked; no-op */
    }
  }

  return (
    <div className="max-w-lg overflow-hidden rounded-[10px] border border-border bg-bg-panel">
      <div
        className="flex flex-wrap gap-1 border-b border-border bg-bg-soft p-1.5"
        role="tablist"
        aria-label="Package manager"
      >
        {MANAGERS.map((m) => (
          <button
            key={m.id}
            role="tab"
            aria-selected={active === m.id}
            onClick={() => setActive(m.id)}
            className={`relative cursor-pointer rounded-md px-3 py-1.5 font-mono text-[13px] transition-colors ${
              active === m.id
                ? "text-accent"
                : "text-faint hover:text-dim"
            }`}
          >
            {active === m.id && (
              <motion.span
                layoutId="install-pill"
                className="absolute inset-0 rounded-md border border-accent-soft bg-bg-panel"
                transition={{ type: "spring", stiffness: 500, damping: 35 }}
              />
            )}
            <span className="relative">{m.label}</span>
          </button>
        ))}
      </div>
      <div className="flex items-start justify-between gap-3 px-4 py-3.5 text-sm">
        <code className="min-w-0 flex-1 font-mono leading-relaxed">
          {current.note && (
            <span className="block text-faint">
              <span className="mr-1.5 text-term">$</span>
              {current.note}
            </span>
          )}
          <span className="block">
            <span className="mr-1.5 text-term">$</span>
            {current.cmd}
          </span>
        </code>
        <button
          onClick={copy}
          aria-label="Copy install command"
          className="shrink-0 cursor-pointer rounded-md border border-border bg-bg-soft px-3 py-1.5 font-mono text-xs text-dim transition-colors hover:border-accent-soft hover:text-accent"
        >
          {copied ? "copied ✓" : "copy"}
        </button>
      </div>
    </div>
  );
}
