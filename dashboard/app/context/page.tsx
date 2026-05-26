"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import type { ContextAnalysis, Event } from "@/lib/types";

export default function ContextPage() {
  const [events, setEvents] = useState<Event[]>([]);
  const [analysis, setAnalysis] = useState<ContextAnalysis | null>(null);

  useEffect(() => {
    api.events(200).then(setEvents).catch(() => {});
    api.contextAnalysis().then(setAnalysis).catch(() => {});
  }, []);

  const llmEvents = events.filter((e) => e.type === "llm_call");

  return (
    <div>
      <h1 className="text-2xl font-semibold mb-6">Context Inspector</h1>

      {analysis && analysis.llm_call_count > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
          <div className="card p-4">
            <div className="text-xs text-zinc-500 mb-1">Prompt tokens</div>
            <div className="text-xl font-semibold">{analysis.total_prompt_tokens.toLocaleString()}</div>
          </div>
          <div className="card p-4">
            <div className="text-xs text-zinc-500 mb-1">LLM calls</div>
            <div className="text-xl font-semibold">{analysis.llm_call_count}</div>
          </div>
          <div className="card p-4">
            <div className="text-xs text-zinc-500 mb-1">Avg efficiency</div>
            <div className="text-xl font-semibold">
              {analysis.avg_efficiency > 0 ? `${analysis.avg_efficiency.toFixed(0)}%` : "—"}
            </div>
          </div>
        </div>
      )}

      {analysis && (analysis.callouts?.length ?? 0) > 0 && (
        <div className="space-y-2 mb-6">
          {analysis.callouts.map((callout) => (
            <div key={callout} className="card p-4 border-amber-500/30 bg-amber-500/5 text-sm text-amber-100">
              {callout}
            </div>
          ))}
        </div>
      )}

      {analysis && ((analysis.duplicate_files?.length ?? 0) > 0 || (analysis.noisy_files?.length ?? 0) > 0) && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4 mb-6">
          {(analysis.duplicate_files?.length ?? 0) > 0 && (
            <div className="card p-4">
              <h2 className="text-sm font-medium text-zinc-300 mb-3">Duplicate context</h2>
              <div className="space-y-2 text-xs">
                {analysis.duplicate_files?.map((f) => (
                  <div key={f.path} className="flex justify-between gap-4">
                    <span className="text-zinc-400 truncate">{f.path}</span>
                    <span className="text-zinc-500 shrink-0">
                      {f.occurrences}× · {f.total_tokens.toLocaleString()} tok
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}
          {(analysis.noisy_files?.length ?? 0) > 0 && (
            <div className="card p-4">
              <h2 className="text-sm font-medium text-zinc-300 mb-3">Noisy context</h2>
              <div className="space-y-2 text-xs">
                {analysis.noisy_files?.map((f) => (
                  <div key={f.path} className="flex justify-between gap-4">
                    <span className="text-zinc-400 truncate">{f.path}</span>
                    <span className="text-red-400 shrink-0">
                      {f.avg_efficiency.toFixed(0)}% · {f.total_tokens.toLocaleString()} tok
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

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
