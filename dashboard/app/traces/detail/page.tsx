"use client";

import { Suspense, useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import { api } from "@/lib/api";
import type { Event } from "@/lib/types";
import { EventRow } from "@/components/ActivityFeed";
import { RestoreButton } from "@/components/RestoreButton";

function TraceDetailInner() {
  const params = useSearchParams();
  const id = params.get("id") || "";
  const [events, setEvents] = useState<Event[]>([]);

  useEffect(() => {
    if (id) api.replay(id).then(setEvents).catch(() => {});
  }, [id]);

  if (!id) {
    return <p className="text-zinc-500">Missing session id.</p>;
  }

  return (
    <div>
      <h1 className="text-2xl font-semibold mb-2">Trace Replay</h1>
      <p className="text-zinc-500 text-sm font-mono mb-6">{id}</p>
      <div className="space-y-3">
        {events.map((e) => (
          <div key={e.id} className="relative group">
            <EventRow event={e} />
            <div className="absolute right-4 top-4 opacity-0 group-hover:opacity-100 transition-opacity">
              <RestoreButton sessionId={id} timestamp={e.timestamp} />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export default function TraceDetailPage() {
  return (
    <Suspense fallback={<p className="text-zinc-500">Loading trace…</p>}>
      <TraceDetailInner />
    </Suspense>
  );
}
