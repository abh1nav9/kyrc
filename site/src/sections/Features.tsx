import { motion } from "motion/react";
import { Reveal } from "../components/Reveal";

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

export function Features() {
  return (
    <section className="grid grid-cols-1 gap-5 py-12 md:grid-cols-2">
      {FEATURES.map((f, i) => (
        <Reveal key={f.title} delay={i * 0.08}>
          <motion.div
            whileHover={{ y: -3 }}
            transition={{ type: "spring", stiffness: 300, damping: 20 }}
            className="h-full rounded-[10px] border border-border bg-bg-panel p-6 transition-colors hover:border-accent-soft"
          >
            <h3 className="mb-2 text-[17px] font-semibold">{f.title}</h3>
            <p className="text-[14.5px] leading-relaxed text-dim">{f.body}</p>
          </motion.div>
        </Reveal>
      ))}
    </section>
  );
}
