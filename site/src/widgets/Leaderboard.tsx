import { useEffect, useState } from "react";

// The public leaderboard, fetched live from the kyrc API. The API base is
// configurable at build time via VITE_LEADERBOARD_URL so the same code works
// against a local server or the hosted one; it falls back to the default.
const API =
  (import.meta.env.VITE_LEADERBOARD_URL as string | undefined) ??
  "https://kyrc-server.onrender.com";

type Entry = {
  rank: number;
  name: string;
  user_id: string;
  wpm: number;
  accuracy: number;
  achieved_at: number;
};

type State =
  | { status: "loading" }
  | { status: "error"; message: string }
  | { status: "ok"; entries: Entry[] };

export function Leaderboard() {
  const [state, setState] = useState<State>({ status: "loading" });

  useEffect(() => {
    let cancelled = false;
    const ctrl = new AbortController();
    const timeout = window.setTimeout(() => ctrl.abort(), 8000);

    fetch(`${API}/leaderboard?limit=20`, { signal: ctrl.signal })
      .then((r) => {
        if (!r.ok) throw new Error(`status ${r.status}`);
        return r.json();
      })
      .then((data: { leaderboard: Entry[] }) => {
        if (!cancelled) setState({ status: "ok", entries: data.leaderboard ?? [] });
      })
      .catch((e) => {
        if (!cancelled)
          setState({
            status: "error",
            message: e?.name === "AbortError" ? "timed out" : String(e?.message ?? e),
          });
      })
      .finally(() => window.clearTimeout(timeout));

    return () => {
      cancelled = true;
      ctrl.abort();
    };
  }, []);

  return (
    <div className="overflow-hidden rounded-[10px] border border-border bg-bg-panel">
      <div className="flex items-center justify-between border-b border-border bg-bg-soft px-4 py-3">
        <span className="font-mono text-sm text-dim">top typists · live</span>
        <span className="font-mono text-xs text-faint">best wpm per player</span>
      </div>

      {state.status === "loading" && (
        <Placeholder text="loading leaderboard…" />
      )}

      {state.status === "error" && (
        <Placeholder
          text="Couldn't reach the leaderboard right now. It syncs when players are online — check back soon."
          sub={`(${state.message})`}
        />
      )}

      {state.status === "ok" && state.entries.length === 0 && (
        <Placeholder text="No scores yet — be the first. Run kyrc, then it syncs your best automatically." />
      )}

      {state.status === "ok" && state.entries.length > 0 && (
        <div className="overflow-x-auto">
          <table className="w-full min-w-[420px] text-sm">
            <thead>
              <tr className="text-left font-mono text-xs text-faint">
                <th className="px-4 py-2 font-normal">#</th>
                <th className="px-4 py-2 font-normal">player</th>
                <th className="px-4 py-2 text-right font-normal">wpm</th>
                <th className="px-4 py-2 text-right font-normal">acc</th>
              </tr>
            </thead>
            <tbody>
              {state.entries.map((e) => (
                <tr
                  key={e.user_id}
                  className="border-t border-border/60 transition-colors hover:bg-bg-soft/40"
                >
                  <td className="px-4 py-2.5 font-mono text-faint">{e.rank}</td>
                  <td className="px-4 py-2.5">
                    <div className="font-medium">{e.name}</div>
                    <div className="font-mono text-[11px] text-faint">{e.user_id}</div>
                  </td>
                  <td className="px-4 py-2.5 text-right font-mono font-bold text-accent">
                    {e.wpm.toFixed(1)}
                  </td>
                  <td className="px-4 py-2.5 text-right font-mono text-dim">
                    {Math.round(e.accuracy * 100)}%
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

function Placeholder({ text, sub }: { text: string; sub?: string }) {
  return (
    <div className="px-4 py-10 text-center">
      <p className="text-sm text-dim">{text}</p>
      {sub && <p className="mt-1 font-mono text-xs text-faint">{sub}</p>}
    </div>
  );
}
