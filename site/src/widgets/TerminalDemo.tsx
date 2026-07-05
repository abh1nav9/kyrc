import { useEffect, useRef, useState } from "react";
import { AnimatePresence, motion } from "motion/react";

// TerminalDemo replays a kyrc session in the browser using the same visual
// language as the real TUI: dim untyped text, bright correct chars, a
// highlighted caret, and a red mistype — then a results panel. It's a
// scripted animation, not the real engine, but it mirrors the states so the
// landing page shows exactly what users will feel.

const TARGET = "the quick brown fox jumps over the lazy dog";

type Step = { type: "hit" } | { type: "miss"; wrong: string } | { type: "back" };

function buildScript(): Step[] {
  const steps: Step[] = [];
  for (let i = 0; i < TARGET.length; i++) {
    // Fumble the 'q' in "quick" once: type 'w', backspace, then correct.
    if (TARGET[i] === "q" && TARGET.slice(i, i + 5) === "quick") {
      steps.push({ type: "miss", wrong: "w" });
      steps.push({ type: "back" });
    }
    steps.push({ type: "hit" });
  }
  return steps;
}

const SCRIPT = buildScript();

type CharState = "untyped" | "correct" | "wrong" | "caret";

export function TerminalDemo() {
  const [typed, setTyped] = useState(0);
  const [wrongChar, setWrongChar] = useState<string | null>(null);
  const [finished, setFinished] = useState(false);
  const stepRef = useRef(0);

  useEffect(() => {
    let timer: number;

    function advance() {
      const i = stepRef.current;
      if (i >= SCRIPT.length) {
        setFinished(true);
        timer = window.setTimeout(() => {
          setFinished(false);
          setTyped(0);
          setWrongChar(null);
          stepRef.current = 0;
          timer = window.setTimeout(advance, 700);
        }, 2600);
        return;
      }

      const step = SCRIPT[i];
      stepRef.current = i + 1;

      if (step.type === "hit") {
        setWrongChar(null);
        setTyped((t) => t + 1);
      } else if (step.type === "miss") {
        setWrongChar(step.wrong);
      } else if (step.type === "back") {
        setWrongChar(null);
      }

      const base = step.type === "hit" ? 55 : step.type === "miss" ? 140 : 180;
      const jitter = Math.random() * 45;
      timer = window.setTimeout(advance, base + jitter);
    }

    timer = window.setTimeout(advance, 700);
    return () => window.clearTimeout(timer);
  }, []);

  function stateFor(i: number): CharState {
    if (i < typed) return "correct";
    if (i === typed) return wrongChar ? "wrong" : "caret";
    return "untyped";
  }

  const charClass: Record<CharState, string> = {
    untyped: "text-untyped",
    correct: "text-text",
    wrong: "text-wrong underline",
    caret: "text-[#0a0c0f] bg-accent rounded-[2px] caret-blink",
  };

  return (
    <div
      className="overflow-hidden rounded-xl border border-border bg-[#0a0c0f] shadow-[0_24px_60px_-20px_rgba(0,0,0,0.8)]"
      role="img"
      aria-label="Animated demo of the kyrc terminal typing test"
    >
      <div className="flex items-center gap-2 border-b border-border bg-bg-soft px-3.5 py-3">
        <span className="inline-block h-[11px] w-[11px] rounded-full bg-[#ff5f56]" />
        <span className="inline-block h-[11px] w-[11px] rounded-full bg-[#ffbd2e]" />
        <span className="inline-block h-[11px] w-[11px] rounded-full bg-[#27c93f]" />
        <span className="ml-2 font-mono text-xs text-faint">kyrc</span>
      </div>

      <div className="flex min-h-[230px] flex-col justify-center px-4 py-7 font-mono sm:px-6">
        <AnimatePresence mode="wait">
          {finished ? (
            <motion.div
              key="results"
              initial={{ opacity: 0, scale: 0.96 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0, scale: 0.96 }}
              transition={{ duration: 0.25 }}
            >
              <Results />
            </motion.div>
          ) : (
            <motion.div
              key="test"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.15 }}
            >
              <div className="mb-[18px] text-[15px] font-bold text-accent">2.4s</div>
              <p className="text-[15px] leading-[1.8] tracking-[0.5px] break-words whitespace-pre-wrap sm:text-[18px]">
                {TARGET.split("").map((ch, i) => {
                  const st = stateFor(i);
                  const shown =
                    st === "wrong" && wrongChar ? (ch === " " ? "_" : ch) : ch;
                  return (
                    <span key={i} className={charClass[st]}>
                      {shown}
                    </span>
                  );
                })}
              </p>
              <div className="mt-[22px] text-[13px] text-faint">
                tab restart · esc quit
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </div>
  );
}

function Results() {
  const rows: [string, string][] = [
    ["wpm", "97"],
    ["raw", "101"],
    ["acc", "98%"],
    ["consistency", "92%"],
    ["time", "5.3s"],
    ["chars", "43/44"],
  ];
  return (
    <div className="text-center">
      <div className="mb-5 text-[15px] font-bold text-accent">results</div>
      <div className="mb-5 grid grid-cols-3 gap-x-3 gap-y-[18px]">
        {rows.map(([label, val], idx) => (
          <motion.div
            key={label}
            className="flex flex-col gap-0.5"
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 + idx * 0.05, duration: 0.25 }}
          >
            <span className="text-2xl font-bold text-accent">{val}</span>
            <span className="text-xs text-faint">{label}</span>
          </motion.div>
        ))}
      </div>
      <div className="text-[13px] text-faint">tab restart · esc quit</div>
    </div>
  );
}
