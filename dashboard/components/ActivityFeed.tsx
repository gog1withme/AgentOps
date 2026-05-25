import type { Event } from "@/lib/types";

export function EfficiencyBadge({ score, promptTokens }: { score?: number; promptTokens?: number }) {
  if (score == null || score === 0) return null;
  const color =
    score >= 70 ? "bg-emerald-500" : score >= 40 ? "bg-amber-500" : "bg-red-500";
  const referenced = promptTokens ? Math.round((score / 100) * promptTokens) : 0;
  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs text-zinc-950 ${color}`}
      title={
        promptTokens
          ? `${referenced.toLocaleString()} of ${promptTokens.toLocaleString()} prompt tokens referenced in output`
          : `${score.toFixed(0)}% efficiency`
      }
    >
      <span className="h-1.5 w-1.5 rounded-full bg-zinc-950/30" />
      {score.toFixed(0)}%
    </span>
  );
}

export function EventRow({ event }: { event: Event }) {
  return (
    <div className="card p-4 flex items-start gap-4">
      <div className="text-xs text-zinc-500 w-20 shrink-0">
        {new Date(event.timestamp).toLocaleTimeString()}
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 flex-wrap">
          <span className="text-xs font-medium uppercase text-violet-400">{event.type}</span>
          <span className="text-xs text-zinc-500">{event.source}</span>
          {event.type === "llm_call" && (
            <EfficiencyBadge score={event.efficiency_score} promptTokens={event.prompt_tokens} />
          )}
        </div>
        <p className="mt-1 text-sm text-zinc-300 truncate">
          {eventSummary(event)}
        </p>
      </div>
      {event.cost_usd != null && event.cost_usd > 0 && (
        <div className="text-xs text-emerald-400 shrink-0">${event.cost_usd.toFixed(4)}</div>
      )}
    </div>
  );
}

function eventSummary(e: Event): string {
  switch (e.type) {
    case "llm_call":
      return `${e.model || "model"} · ${(e.prompt_tokens || 0) + (e.output_tokens || 0)} tokens`;
    case "file_edit":
      return e.file_path || "file edited";
    case "shell_command":
      return e.shell_command || "shell";
    case "mcp_call":
      return `${e.mcp_server} · ${e.mcp_latency_ms}ms`;
    default:
      return e.tool_name || e.error || e.type;
  }
}
