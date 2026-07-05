import { NPM, REPO } from "../lib/constants";

export function Footer() {
  return (
    <footer className="mt-10 flex flex-col items-center gap-3 border-t border-border py-10 pb-14 text-center font-mono">
      <div className="text-[15px] font-bold">
        <span className="text-accent">⌨</span> kyrc
      </div>
      <div className="flex gap-5 text-sm text-dim">
        <a
          href={REPO}
          target="_blank"
          rel="noreferrer"
          className="transition-colors hover:text-accent"
        >
          github
        </a>
        <a
          href={NPM}
          target="_blank"
          rel="noreferrer"
          className="transition-colors hover:text-accent"
        >
          npm
        </a>
      </div>
      <div className="text-[12.5px] text-faint">
        MIT licensed · built in Go &amp; Bubble Tea · proudly open source
      </div>
    </footer>
  );
}
