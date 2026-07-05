import { Reveal } from "../components/Reveal";
import { Section } from "../components/Section";
import { SectionHeading } from "../components/SectionHeading";
import { Panel } from "../components/Panel";

const FLAGS: [string, string][] = [
  ["kyrc", "25-word test (default)"],
  ["kyrc -w 50", "50-word test"],
  ["kyrc -t 30", "30-second test"],
  ["kyrc -t 1m", "1-minute test"],
  ["kyrc -q", "random quote"],
  ["kyrc results", "your last 10 results (sortable)"],
  ["kyrc login", "create or restore an account"],
  ["kyrc leaderboard", "view the global leaderboard"],
  ["kyrc sync", "push your best result now"],
  ["kyrc update", "update to the latest version"],
];

const KEYS: [string, string][] = [
  ["type", "start the test on first keystroke"],
  ["backspace", "delete a character"],
  ["ctrl+w", "delete the previous word"],
  ["tab", "restart the test"],
  ["esc / ctrl+c", "quit"],
];

export function Docs() {
  return (
    <Section id="usage">
      <Reveal>
        <SectionHeading label="kyrc --help">Usage</SectionHeading>
      </Reveal>
      <div className="grid grid-cols-1 gap-5 md:grid-cols-2">
        <Reveal>
          <Panel label="Commands">
            <table className="w-full border-collapse">
              <tbody>
                {FLAGS.map(([cmd, desc]) => (
                  <tr key={cmd}>
                    <td className="w-[42%] py-1.5 pr-3 align-top">
                      <code className="font-mono text-[13.5px] text-accent">
                        {cmd}
                      </code>
                    </td>
                    <td className="py-1.5 align-top text-sm text-dim">{desc}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Panel>
        </Reveal>
        <Reveal delay={0.08}>
          <Panel label="Keys">
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
          </Panel>
        </Reveal>
      </div>
    </Section>
  );
}
