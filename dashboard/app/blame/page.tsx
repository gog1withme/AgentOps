"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import type { Attribution } from "@/lib/types";

export default function BlamePage() {
  const [file, setFile] = useState("");
  const [rows, setRows] = useState<Attribution[]>([]);

  async function load(path: string) {
    if (!path) return;
    const data = await api.blame(path).catch(() => []);
    setRows(data);
  }

  useEffect(() => {
    if (file) load(file);
  }, [file]);

  return (
    <div>
      <h1 className="text-2xl font-semibold mb-6">AI Blame</h1>
      <input
        className="w-full max-w-lg mb-6 bg-zinc-900 border border-zinc-800 rounded-lg px-4 py-2 text-sm"
        placeholder="Enter file path…"
        value={file}
        onChange={(e) => setFile(e.target.value)}
        onKeyDown={(e) => e.key === "Enter" && load(file)}
      />
      <div className="space-y-3">
        {rows.map((r) => (
          <div key={r.id} className="card p-4">
            <div className="flex items-center gap-2 text-sm">
              <span className="text-zinc-500">{new Date(r.timestamp).toLocaleString()}</span>
              <span className="px-2 py-0.5 rounded bg-violet-500/20 text-violet-300 text-xs">{r.agent}</span>
              <span className="text-emerald-400">+{r.lines_added}</span>
              <span className="text-red-400">-{r.lines_removed}</span>
            </div>
            {r.diff && (
              <pre className="mt-3 text-xs text-zinc-400 overflow-x-auto max-h-48">{r.diff}</pre>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
