import { motion } from "motion/react";
import { InstallTabs } from "../widgets/InstallTabs";
import { TerminalDemo } from "../widgets/TerminalDemo";

export function Hero() {
  return (
    <header
      id="top"
      className="grid grid-cols-1 items-center gap-12 py-12 md:grid-cols-2 md:py-20"
    >
      <motion.div
        initial={{ opacity: 0, y: 24 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.6, ease: "easeOut" }}
      >
        <span className="font-mono text-xs uppercase tracking-[2px] text-accent-soft">
          <span className="text-term">$</span> terminal typing test
        </span>
        <h1 className="my-4 text-[clamp(28px,7vw,52px)] font-extrabold leading-[1.08] tracking-[-1px] text-balance sm:tracking-[-1.5px]">
          Type faster,{" "}
          <span className="text-accent">without leaving the terminal.</span>
        </h1>
        <p className="mb-7 max-w-[30rem] text-[17px] leading-relaxed text-dim">
          kyrc is a fast, offline, keyboard-only typing test that lives where
          you already work. It starts in milliseconds, never asks you to log in,
          and works with no network.
        </p>
        <InstallTabs />
        <div className="mt-4 text-sm text-faint">
          Then just run{" "}
          <code className="rounded bg-bg-soft px-1.5 py-0.5 font-mono text-accent">
            kyrc
          </code>
          . That&apos;s it. Already installed?{" "}
          <code className="rounded bg-bg-soft px-1.5 py-0.5 font-mono text-accent">
            kyrc update
          </code>{" "}
          gets the latest.
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
