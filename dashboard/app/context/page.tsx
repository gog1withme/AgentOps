"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import type { Event } from "@/lib/types";

export default function ContextPage() {
  const [events, setEvents] = useState<Event[]>([]);

  useEffect(() => {
    api.events(200).then(setEvents).catch(() => {});
  }, []);

  const llmEvents = events.filter((e) => e.type === "llm_call");

  return (
    <div>
      <h1 className="text-2xl font-semibold mb-6">Context Inspector</h1>
      <div className="space-y-3">
        {llmEvents.map((e) => (
          <div key={e.id} className="card p-4">
            <div className="flex justify-between text-sm mb-2">
              <span className="text-violet-300">{e.model}</span>
              <span className="text-zinc-500">{e.prompt_tokens?.toLocaleString()} prompt tokens</span>
            </div>
            <div className="text-xs text-zinc-500">
              Efficiency: {e.efficiency_score?.toFixed(0) ?? "—"}% · Hash: {e.prompt_hash?.slice(0, 8)}
            </div>
            {e.file_path && (
              <div className="mt-2 text-xs text-zinc-400">Referenced: {e.file_path}</div>
            )}
          </div>
        ))}
        {llmEvents.length === 0 && (
          <div className="card p-8 text-center text-zinc-500">No LLM calls recorded yet.</div>
        )}
      </div>
    </div>
  );
}
