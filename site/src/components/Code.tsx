import type { ReactNode } from "react";

// Code renders a single shell command with a green `$` prompt, matching the
// terminal aesthetic. Long commands scroll inside the box, never the page.
export function Code({ children }: { children: ReactNode }) {
  return (
    <pre className="overflow-x-auto rounded-md border border-border/60 bg-bg-soft px-3 py-2 font-mono text-[13px] text-text">
      <span className="mr-1.5 text-term">$</span>
      {children}
    </pre>
  );
}
