import { Reveal } from "../components/Reveal";
import { Section } from "../components/Section";
import { SectionHeading } from "../components/SectionHeading";
import { Panel } from "../components/Panel";

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

const TREE = `cmd/kyrc            CLI entry, flag parsing, version metadata
internal/engine     pure state machine + metrics (no terminal)
internal/store      local results history (offline, last 10)
internal/identity   Ed25519 keypair, user_id, recovery phrase
internal/leaderboard signed submission + replay anti-cheat
internal/ui         Bubble Tea adapter — renders, owns no math
server/             leaderboard API in front of Postgres
site/               this landing page`;

export function Architecture() {
  return (
    <Section id="architecture">
      <Reveal>
        <SectionHeading label="kyrc --internals">
          Engineered as a feedback loop, not a form
        </SectionHeading>
      </Reveal>
      <div className="mb-7 grid grid-cols-1 gap-[18px] md:grid-cols-3">
        {ARCH.map((a, i) => (
          <Reveal key={a.title} delay={i * 0.08}>
            <Panel>
              <h3 className="mb-2 text-base font-semibold text-accent">
                {a.title}
              </h3>
              <p className="text-sm leading-relaxed text-dim">{a.body}</p>
            </Panel>
          </Reveal>
        ))}
      </div>
      <Reveal>
        <pre className="overflow-x-auto rounded-[10px] border border-border bg-[#0a0c0f] p-5.5 text-[13px] leading-[1.7] text-dim">
          <code>{TREE}</code>
        </pre>
      </Reveal>
    </Section>
  );
}
