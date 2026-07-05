import { Reveal } from "../components/Reveal";
import { Section } from "../components/Section";
import { SectionHeading } from "../components/SectionHeading";

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

export function Metrics() {
  return (
    <Section id="metrics">
      <Reveal>
        <SectionHeading label="kyrc --metrics">
          How the numbers are defined
        </SectionHeading>
        <p className="mb-7 max-w-[44rem] leading-relaxed text-dim">
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
              <p className="text-[15px] leading-relaxed text-dim">{m.def}</p>
            </div>
          </Reveal>
        ))}
      </div>
    </Section>
  );
}
