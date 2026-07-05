import type { ReactNode } from "react";

// Panel is the shared bordered card used across sections (features, docs,
// account steps, architecture). Centralizing it keeps radius/border/padding
// consistent everywhere and gives one place to tune the surface style.
export function Panel({
  children,
  className = "",
  hover = false,
  label,
}: {
  children: ReactNode;
  className?: string;
  hover?: boolean;
  // Optional monospace header label (uppercase eyebrow) for titled panels.
  label?: string;
}) {
  return (
    <div
      className={`h-full rounded-[10px] border border-border bg-bg-panel p-5.5 transition-colors ${
        hover ? "hover:border-accent-soft" : ""
      } ${className}`}
    >
      {label && (
        <h3 className="mb-3.5 font-mono text-[13px] uppercase tracking-[1.5px] text-accent-soft">
          {label}
        </h3>
      )}
      {children}
    </div>
  );
}

// StepCard is a titled Panel used by the account walkthrough.
export function StepCard({
  title,
  children,
}: {
  title: string;
  children: ReactNode;
}) {
  return (
    <Panel label={title}>
      <div className="space-y-2.5 text-sm text-dim">{children}</div>
    </Panel>
  );
}
