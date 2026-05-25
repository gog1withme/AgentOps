"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import type { SessionSummary } from "@/lib/types";

export default function TracesPage() {
  const [sessions, setSessions] = useState<SessionSummary[]>([]);

  useEffect(() => {
    api.sessions().then(setSessions).catch(() => {});
  }, []);

  return (
    <div>
      <h1 className="text-2xl font-semibold mb-6">Traces</h1>
      <div className="space-y-2">
        {sessions.map((s) => (
          <Link key={s.session_id} href={`/traces/detail/?id=${encodeURIComponent(s.session_id)}`} className="card block p-4 hover:border-violet-500/50">
            <div className="font-mono text-sm text-violet-300">{s.session_id.slice(0, 18)}…</div>
            <div className="text-xs text-zinc-500 mt-1">
              {s.event_count} events · ${s.total_cost.toFixed(4)} · {new Date(s.last_event).toLocaleString()}
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}
