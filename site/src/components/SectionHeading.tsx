import type { ReactNode } from "react";

// SectionHeading pairs a monospace "terminal prompt" eyebrow with the title,
// reinforcing the terminal identity that starts at the ⌨ kyrc logo. The
// `label` renders like a shell command (e.g. `kyrc --leaderboard`).
export function SectionHeading({
  label,
  children,
}: {
  label?: string;
  children: ReactNode;
}) {
  return (
    <div className="mb-5">
      {label && (
        <div className="mb-2.5 font-mono text-[13px] text-accent-soft">
          <span className="text-term">$</span> {label}
        </div>
      )}
      <h2 className="text-[clamp(24px,3.4vw,32px)] font-semibold tracking-[-0.8px] text-balance">
        {children}
      </h2>
    </div>
  );
}
