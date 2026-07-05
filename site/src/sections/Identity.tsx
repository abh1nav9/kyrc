import { Reveal } from "../components/Reveal";
import { Section } from "../components/Section";
import { SectionHeading } from "../components/SectionHeading";
import { StepCard } from "../components/Panel";
import { Code } from "../components/Code";

// Identity documents the account model: creating one, finding your user_id +
// passkey, and logging back in on another machine.
export function Identity() {
  return (
    <Section id="account">
      <Reveal>
        <SectionHeading label="kyrc login">Your account &amp; passkey</SectionHeading>
        <p className="mb-6 max-w-[46rem] leading-relaxed text-dim">
          Accounts are optional — kyrc works with no login. When you want on the
          leaderboard, kyrc gives you a{" "}
          <span className="text-text">user_id</span> and a{" "}
          <span className="text-text">passkey</span> (a recovery phrase). Your
          private key never leaves your machine; the passkey is the only way to
          sign in elsewhere, so keep it safe.
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
              <code className="text-accent">sync</code> forces it. Everything
              else stays fully offline.
            </p>
          </StepCard>
        </Reveal>
      </div>
    </Section>
  );
}
