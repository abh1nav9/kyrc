import { Reveal } from "../components/Reveal";
import { Section } from "../components/Section";
import { SectionHeading } from "../components/SectionHeading";
import { Leaderboard } from "../widgets/Leaderboard";

export function LeaderboardSection() {
  return (
    <Section id="leaderboard">
      <Reveal>
        <SectionHeading label="kyrc leaderboard">Leaderboard</SectionHeading>
        <p className="mb-6 max-w-[46rem] leading-relaxed text-dim">
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
    </Section>
  );
}
