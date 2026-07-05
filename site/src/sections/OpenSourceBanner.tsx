import { REPO } from "../lib/constants";

// A slim strip stating kyrc is open source — a point of pride and trust,
// linking straight to the repo.
export function OpenSourceBanner() {
  return (
    <a
      href={REPO}
      target="_blank"
      rel="noreferrer"
      className="mt-3 flex items-center justify-center gap-2 rounded-lg border border-accent-soft/70 bg-accent-soft/10 px-4 py-2 text-center text-sm text-dim transition-colors hover:border-accent hover:text-text"
    >
      <span className="text-accent">★</span>
      <span>
        We are proudly{" "}
        <span className="font-semibold text-accent">OPEN SOURCE</span> — read
        every line, audit the crypto, send a PR.
      </span>
      <span className="text-faint">↗</span>
    </a>
  );
}
