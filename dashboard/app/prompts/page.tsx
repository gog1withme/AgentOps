"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import type { Prompt } from "@/lib/types";

export default function PromptsPage() {
  const [prompts, setPrompts] = useState<Prompt[]>([]);
  const [selected, setSelected] = useState<Prompt | null>(null);
  const [compareA, setCompareA] = useState("");
  const [compareB, setCompareB] = useState("");
  const [diff, setDiff] = useState("");

  useEffect(() => {
    api.prompts().then(setPrompts).catch(() => {});
  }, []);

  async function runDiff() {
    if (!compareA || !compareB) return;
    const res = await fetch(
      `${typeof window !== "undefined" ? window.location.origin : "http://127.0.0.1:4318"}/api/prompts/diff?a=${compareA}&b=${compareB}`
    );
    const data = await res.json();
    setDiff(data.diff || "");
  }

  return (
    <div>
      <h1 className="text-2xl font-semibold mb-6">Prompt Versions</h1>
      <div className="grid lg:grid-cols-2 gap-6">
        <div className="card overflow-hidden">
          <table className="w-full text-sm">
            <thead className="border-b border-zinc-800 text-zinc-500">
              <tr>
                <th className="text-left p-3">Hash</th>
                <th className="text-right p-3">Sessions</th>
                <th className="text-right p-3">Avg cost</th>
                <th className="text-right p-3">Eff.</th>
              </tr>
            </thead>
            <tbody>
              {prompts.map((p) => (
                <tr
                  key={p.hash}
                  className="border-b border-zinc-800/50 cursor-pointer hover:bg-zinc-800/30"
                  onClick={() => setSelected(p)}
                >
                  <td className="p-3 font-mono text-violet-300">{p.hash.slice(0, 8)}</td>
                  <td className="p-3 text-right">{p.session_count}</td>
                  <td className="p-3 text-right">${p.avg_cost_usd.toFixed(4)}</td>
                  <td className="p-3 text-right">{p.avg_efficiency_score.toFixed(0)}%</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <div className="card p-4">
          {selected ? (
            <pre className="text-xs text-zinc-400 whitespace-pre-wrap max-h-96 overflow-auto">{selected.content}</pre>
          ) : (
            <p className="text-zinc-500 text-sm">Select a prompt to view content</p>
          )}
          <div className="mt-6 border-t border-zinc-800 pt-4">
            <h3 className="text-sm font-medium mb-2">Compare</h3>
            <div className="flex gap-2 mb-2">
              <input className="flex-1 bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs" placeholder="hash A" value={compareA} onChange={(e) => setCompareA(e.target.value)} />
              <input className="flex-1 bg-zinc-950 border border-zinc-800 rounded px-2 py-1 text-xs" placeholder="hash B" value={compareB} onChange={(e) => setCompareB(e.target.value)} />
              <button onClick={runDiff} className="px-3 py-1 text-xs rounded bg-violet-500">Diff</button>
            </div>
            {diff && <pre className="text-xs text-zinc-400 whitespace-pre-wrap">{diff}</pre>}
          </div>
        </div>
      </div>
    </div>
  );
}
