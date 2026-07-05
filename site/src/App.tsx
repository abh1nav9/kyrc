import { motion } from "motion/react";
import type { ReactNode } from "react";
import { TerminalDemo } from "./TerminalDemo";
import { InstallTabs } from "./InstallTabs";
import { Leaderboard } from "./Leaderboard";

const REPO = "https://github.com/abh1nav9/kyrc";

// Reveal fades + lifts its children into view once, on scroll. Used to give
// each section a subtle entrance without a heavy animation library setup.
function Reveal({
  children,
  delay = 0,
  className = "",
}: {
  children: ReactNode;
  delay?: number;
  className?: string;
}) {
  return (
    <motion.div
      className={className}
      initial={{ opacity: 0, y: 20 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, margin: "-80px" }}
      transition={{ duration: 0.5, delay, ease: "easeOut" }}
    >
      {children}
    </motion.div>
  );
}

export function App() {
  return (
    <div className="relative z-[1] mx-auto max-w-[1080px] px-6">
      <Nav />
      <OpenSourceBanner />
      <Hero />
      <Features />
      <LeaderboardSection />
      <Docs />
      <Identity />
      <Metrics />
      <Architecture />
      <Footer />
    </div>
  );
}

// OpenSourceBanner is a slim, dismissible-free strip stating kyrc is open
// source — a point of pride and trust, linking straight to the repo.
function OpenSourceBanner() {
  return (
    <a
      href={REPO}
      target="_blank"
      rel="noreferrer"
      className="mt-3 flex items-center justify-center gap-2 rounded-lg border border-accent-soft bg-accent-soft/10 px-4 py-2 text-center text-sm text-dim transition-colors hover:border-accent hover:text-text"
    >
      <span className="text-accent">★</span>
      <span>
        We are proudly <span className="font-semibold text-accent">OPEN SOURCE</span> — read
        every line, audit the crypto, send a PR.
      </span>
      <span className="text-faint">↗</span>
    </a>
  );
}

function LeaderboardSection() {
  return (
    <section id="leaderboard" className="border-t border-border py-14">
      <Reveal>
        <SectionHeading>Leaderboard</SectionHeading>
        <p className="mb-6 max-w-[46rem] text-dim">
          kyrc runs fully offline — but log in with a name and your best result
          syncs to a global leaderboard whenever you're online. Scores are{" "}
          <span className="text-text">signed on your device</span> and{" "}
          <span className="text-text">replayed on the server</span> from the raw
          keystroke log, so nobody can fake a WPM or submit as someone else.
        </p>
      </Reveal>
      <Reveal delay={0.05}>
        <Leaderboard />
      </Reveal>
    </section>
  );
}

function Nav() {
  return (
    <nav className="flex items-center justify-between border-b border-border py-5.5">
      <a href="#top" className="font-mono text-lg font-bold tracking-tight">
        <span className="text-accent">⌨</span> kyrc
      </a>
      <div className="flex gap-[22px] text-sm text-dim">
        <a href="#leaderboard" className="transition-colors hover:text-text">leaderboard</a>
        <a href="#usage" className="transition-colors hover:text-text">docs</a>
        <a href="#account" className="transition-colors hover:text-text">account</a>
        <a href="#architecture" className="transition-colors hover:text-text">architecture</a>
        <a href={REPO} target="_blank" rel="noreferrer" className="transition-colors hover:text-text">
          github ↗
        </a>
      </div>
    </nav>
  );
}

function Hero() {
  return (
    <header id="top" className="grid grid-cols-1 items-center gap-12 py-12 md:grid-cols-2 md:py-18">
      <motion.div
        initial={{ opacity: 0, y: 24 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, ease: "easeOut" }}
      >
        <span className="font-mono text-xs uppercase tracking-[2px] text-accent-soft">
          terminal typing test
        </span>
        <h1 className="my-4 text-[clamp(34px,5vw,52px)] font-extrabold leading-[1.1] tracking-[-1.5px]">
          Type faster,
          <br />
          <span className="text-accent">without leaving the terminal.</span>
        </h1>
        <p className="mb-7 max-w-[30rem] text-[17px] text-dim">
          kyrc is a fast, offline, keyboard-only typing test that lives where
          you already work. It starts in milliseconds, never asks you to log
          in, and works with no network.
        </p>
        <InstallTabs />
        <div className="mt-4 text-sm text-faint">
          Then just run{" "}
          <code className="rounded bg-bg-soft px-1.5 py-0.5 font-mono text-accent">kyrc</code>
          . That&apos;s it.
        </div>
      </motion.div>
      <motion.div
        initial={{ opacity: 0, scale: 0.96 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ duration: 0.6, delay: 0.15, ease: "easeOut" }}
      >
        <TerminalDemo />
      </motion.div>
    </header>
  );
}

const FEATURES: { title: string; body: string }[] = [
  {
    title: "Instant, always",
    body: "A static Go binary with no runtime. First keystroke to first frame is engineered to feel free — feedback lands in a single frame.",
  },
  {
    title: "Offline & private",
    body: "No network, no account, no telemetry. Practice on a plane, in a container, over SSH. Your keystrokes never leave your machine.",
  },
  {
    title: "Honest metrics",
    body: "WPM, accuracy, and consistency computed from a timestamped keystroke log — reproducible and auditable, matching Monkeytype's conventions.",
  },
  {
    title: "Built for the loop",
    body: "The clock starts on your first keystroke, pastes are rejected, and typing feedback is decoupled from the clock so nothing ever feels laggy.",
  },
];

function Features() {
  return (
    <section className="grid grid-cols-1 gap-5 py-10 md:grid-cols-2">
      {FEATURES.map((f, i) => (
        <Reveal key={f.title} delay={i * 0.08}>
          <motion.div
            whileHover={{ y: -2 }}
            className="h-full rounded-[10px] border border-border bg-bg-panel p-6 transition-colors hover:border-accent-soft"
          >
            <h3 className="mb-2 text-[17px] font-semibold">{f.title}</h3>
            <p className="text-[14.5px] text-dim">{f.body}</p>
          </motion.div>
        </Reveal>
      ))}
    </section>
  );
}

const FLAGS: [string, string][] = [
  ["kyrc", "25-word test (default)"],
  ["kyrc -w 50", "50-word test"],
  ["kyrc -t 30", "30-second test"],
  ["kyrc -t 1m", "1-minute test"],
  ["kyrc -q", "random quote"],
];

const KEYS: [string, string][] = [
  ["type", "start the test on first keystroke"],
  ["backspace", "delete a character"],
  ["ctrl+w", "delete the previous word"],
  ["tab", "restart the test"],
  ["esc / ctrl+c", "quit"],
];

function SectionHeading({ children }: { children: ReactNode }) {
  return (
    <h2 className="mb-3 text-[clamp(24px,3.4vw,32px)] tracking-[-0.8px]">
      {children}
    </h2>
  );
}

function Docs() {
  return (
    <section id="usage" className="border-t border-border py-14">
      <Reveal>
        <SectionHeading>Usage</SectionHeading>
      </Reveal>
      <div className="mt-6 grid grid-cols-1 gap-5 md:grid-cols-2">
        <Reveal>
          <div className="rounded-[10px] border border-border bg-bg-panel p-5.5">
            <h3 className="mb-3.5 font-mono text-[13px] uppercase tracking-[1.5px] text-accent-soft">
              Commands
            </h3>
            <table className="w-full border-collapse">
              <tbody>
                {FLAGS.map(([cmd, desc]) => (
                  <tr key={cmd}>
                    <td className="w-[42%] py-1.5 pr-3 align-top">
                      <code className="font-mono text-[13.5px] text-accent">{cmd}</code>
                    </td>
                    <td className="py-1.5 align-top text-sm text-dim">{desc}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Reveal>
        <Reveal delay={0.08}>
          <div className="rounded-[10px] border border-border bg-bg-panel p-5.5">
            <h3 className="mb-3.5 font-mono text-[13px] uppercase tracking-[1.5px] text-accent-soft">
              Keys
            </h3>
            <table className="w-full border-collapse">
              <tbody>
                {KEYS.map(([key, desc]) => (
                  <tr key={key}>
                    <td className="w-[42%] py-1.5 pr-3 align-top">
                      <kbd className="rounded border border-border border-b-2 bg-bg-soft px-1.5 py-0.5 font-mono text-[12.5px] text-text">
                        {key}
                      </kbd>
                    </td>
                    <td className="py-1.5 align-top text-sm text-dim">{desc}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Reveal>
      </div>
    </section>
  );
}

// Identity documents the account model: creating one, finding your user_id +
// passkey, and logging back in on another machine.
function Identity() {
  return (
    <section id="account" className="border-t border-border py-14">
      <Reveal>
        <SectionHeading>Your account &amp; passkey</SectionHeading>
        <p className="mb-6 max-w-[46rem] text-dim">
          Accounts are optional — kyrc works with no login. When you want on the
          leaderboard, kyrc gives you a <span className="text-text">user_id</span>{" "}
          and a <span className="text-text">passkey</span> (a recovery phrase).
          Your private key never leaves your machine; the passkey is the only
          way to sign in elsewhere, so keep it safe.
        </p>
      </Reveal>

      <div className="grid grid-cols-1 gap-5 md:grid-cols-2">
        <Reveal>
          <StepCard title="Create your account">
            <Code>{`kyrc login "Your Name"`}</Code>
            <p>
              Prints your <b>user_id</b> and <b>passkey</b>, and saves a private
              recovery card locally. Write the passkey down.
            </p>
          </StepCard>
        </Reveal>

        <Reveal delay={0.05}>
          <StepCard title="See your user_id & passkey anytime">
            <Code>kyrc whoami</Code>
            <p>Shows your details and the exact path to your recovery file:</p>
            <ul className="mt-1 space-y-1 font-mono text-[12.5px] text-faint">
              <li>macOS · ~/Library/Application Support/kyrc/recovery.txt</li>
              <li>Linux · ~/.config/kyrc/recovery.txt</li>
              <li>Windows · %AppData%\kyrc\recovery.txt</li>
            </ul>
          </StepCard>
        </Reveal>

        <Reveal delay={0.1}>
          <StepCard title="Log in again on a new machine">
            <Code>kyrc login</Code>
            <p>
              Choose <b>“restore”</b>, then enter your <b>user_id</b> and{" "}
              <b>passkey</b>. kyrc rebuilds the same account — same user_id, same
              scores.
            </p>
          </StepCard>
        </Reveal>

        <Reveal delay={0.15}>
          <StepCard title="Push your best & view the board">
            <Code>kyrc sync</Code>
            <Code>kyrc leaderboard</Code>
            <p>
              Your best result auto-syncs after each test when you're online;{" "}
              <code className="text-accent">sync</code> forces it. Everything else
              stays fully offline.
            </p>
          </StepCard>
        </Reveal>
      </div>
    </section>
  );
}

function StepCard({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="h-full rounded-[10px] border border-border bg-bg-panel p-5.5">
      <h3 className="mb-3 font-mono text-[13px] uppercase tracking-[1.5px] text-accent-soft">
        {title}
      </h3>
      <div className="space-y-2.5 text-sm text-dim">{children}</div>
    </div>
  );
}

function Code({ children }: { children: ReactNode }) {
  return (
    <pre className="overflow-x-auto rounded-md bg-bg-soft px-3 py-2 font-mono text-[13px] text-text">
      <span className="mr-1.5 text-term">$</span>
      {children}
    </pre>
  );
}

const METRICS: { name: string; def: string }[] = [
  {
    name: "wpm",
    def: "correct characters ÷ 5 ÷ minutes — the standard 5-chars-per-word convention. This is the hero number, matching Monkeytype's headline.",
  },
  {
    name: "raw",
    def: "all typed characters ÷ 5 ÷ minutes, ignoring correctness. How fast your fingers moved, mistakes included.",
  },
  {
    name: "acc",
    def: "correct keystrokes ÷ total keystrokes. Keystroke accuracy — a char you mistyped and fixed still counts as an error even if the final text is clean.",
  },
  {
    name: "consistency",
    def: "1 − coefficient of variation of your per-second raw WPM. Higher means steadier pace.",
  },
];

function Metrics() {
  return (
    <section id="metrics" className="border-t border-border py-14">
      <Reveal>
        <SectionHeading>How the numbers are defined</SectionHeading>
        <p className="mb-7 max-w-[44rem] text-dim">
          Users compare their number to Monkeytype and rage if it&apos;s off, so
          every metric is pinned to a precise definition and computed as a pure
          function of the keystroke log. The clock starts on your{" "}
          <strong className="text-text">first keystroke</strong> — idle time
          never counts — and pasting is rejected so it can&apos;t inflate WPM.
        </p>
      </Reveal>
      <div className="flex flex-col">
        {METRICS.map((m, i) => (
          <Reveal key={m.name} delay={i * 0.05}>
            <div
              className={`grid grid-cols-1 items-baseline gap-x-5 gap-y-1.5 border-t border-border py-[18px] md:grid-cols-[140px_1fr] ${
                i === METRICS.length - 1 ? "border-b" : ""
              }`}
            >
              <code className="text-base font-bold text-accent">{m.name}</code>
              <p className="text-[15px] text-dim">{m.def}</p>
            </div>
          </Reveal>
        ))}
      </div>
    </section>
  );
}

const ARCH: { title: string; body: string }[] = [
  {
    title: "Own the clock",
    body: "Every keystroke is timestamped at capture. The engine is a pure function of that timestamped event stream — no clock is ever read inside the logic — so any run is fully reproducible and replayable.",
  },
  {
    title: "Keep the engine pure",
    body: "The typing engine is a headless finite state machine (idle → running → finished) with metrics as pure functions over an append-only keystroke log. It's unit-tested with synthetic timestamps — the terminal UI is a thin adapter on top.",
  },
  {
    title: "Make input→pixel feel free",
    body: "Per-keystroke feedback is immediate; the countdown redraws at ~15Hz. A late clock frame is invisible; a late keystroke is not. The render path stays allocation-light to dodge GC hitches.",
  },
];

function Architecture() {
  return (
    <section id="architecture" className="border-t border-border py-14">
      <Reveal>
        <SectionHeading>Engineered as a feedback loop, not a form</SectionHeading>
      </Reveal>
      <div className="my-7 grid grid-cols-1 gap-[18px] md:grid-cols-3">
        {ARCH.map((a, i) => (
          <Reveal key={a.title} delay={i * 0.08}>
            <div className="h-full rounded-[10px] border border-border bg-bg-panel p-5.5">
              <h3 className="mb-2 text-base font-semibold text-accent">{a.title}</h3>
              <p className="text-sm text-dim">{a.body}</p>
            </div>
          </Reveal>
        ))}
      </div>
      <Reveal>
        <pre className="overflow-x-auto rounded-[10px] border border-border bg-[#0a0c0f] p-5.5 text-[13px] leading-[1.7] text-dim">
          <code>{`cmd/kyrc            CLI entry, flag parsing, version metadata
internal/engine     pure state machine + metrics (no terminal)
internal/wordsource word / quote generation behind an interface
internal/input      terminal key → engine event (timestamped)
internal/ui         Bubble Tea adapter — renders, owns no math
npm/                npm distribution (meta pkg + platform pkgs)
site/               this landing page`}</code>
        </pre>
      </Reveal>
    </section>
  );
}

function Footer() {
  return (
    <footer className="flex flex-col items-center gap-3 border-t border-border py-10 pb-14 text-center font-mono">
      <div>
        <span className="text-accent">⌨</span> kyrc
      </div>
      <div className="flex gap-5 text-sm text-dim">
        <a href={REPO} target="_blank" rel="noreferrer" className="transition-colors hover:text-accent">
          GitHub
        </a>
        <a href="https://www.npmjs.com/package/@kyrc/kyrc" target="_blank" rel="noreferrer" className="transition-colors hover:text-accent">
          npm
        </a>
      </div>
      <div className="text-[12.5px] text-faint">
        MIT licensed · built in Go &amp; Bubble Tea
      </div>
    </footer>
  );
}
