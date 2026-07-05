import { NAV_LINKS, REPO } from "../lib/constants";

// Nav is styled to match the ⌨ kyrc logo: everything is monospace and reads
// like a terminal. Links render as `--flags`, so the whole bar feels like one
// coherent command line rather than a generic website header.
export function Nav() {
  return (
    <nav className="sticky top-0 z-50 -mx-4 flex items-center justify-between gap-3 border-b border-border bg-bg/80 px-4 py-3.5 font-mono backdrop-blur-md sm:-mx-6 sm:px-6">
      <a
        href="#top"
        className="text-lg font-bold tracking-tight transition-colors hover:text-accent"
      >
        <span className="text-accent">⌨</span> kyrc
      </a>

      {/* Middle links are hidden on small screens; the Star CTA stays. */}
      <div className="hidden items-center gap-1 text-[13px] text-dim md:flex">
        {NAV_LINKS.map((l) => (
          <a
            key={l.href}
            href={l.href}
            className="rounded px-2.5 py-1.5 transition-colors hover:bg-bg-soft hover:text-accent"
          >
            <span className="text-faint">--</span>
            {l.label}
          </a>
        ))}
      </div>

      <StarButton />
    </nav>
  );
}

// StarButton is the "star the repo" CTA — a GitHub link styled as a terminal
// button, monospace to match the nav.
export function StarButton() {
  return (
    <a
      href={REPO}
      target="_blank"
      rel="noreferrer"
      className="flex shrink-0 items-center gap-1.5 rounded-md border border-border bg-bg-soft px-3 py-1.5 text-[13px] text-dim transition-colors hover:border-accent-soft hover:text-accent"
    >
      <span className="text-accent">★</span>
      <span>star</span>
      <span className="hidden text-faint sm:inline">· github</span>
    </a>
  );
}
