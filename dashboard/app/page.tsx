"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import type { Event } from "@/lib/types";
import { EventRow } from "@/components/ActivityFeed";

export default function HomePage() {
  const [events, setEvents] = useState<Event[]>([]);

  useEffect(() => {
    api.events(50).then(setEvents).catch(() => {});
    const es = new EventSource(api.streamURL());
    es.onmessage = (msg) => {
      try {
        const payload = JSON.parse(msg.data);
        if (payload.type === "event" && payload.data) {
          setEvents((prev) => [payload.data as Event, ...prev].slice(0, 100));
        }
      } catch {
        /* ignore */
      }
    };
    return () => es.close();
  }, []);

  return (
    <div>
      <h1 className="text-2xl font-semibold mb-1">Live Activity</h1>
      <p className="text-zinc-500 text-sm mb-6">Real-time stream from your AI agents</p>
      <div className="space-y-3">
        {events.length === 0 && (
          <div className="card p-8 text-center text-zinc-500">
            Waiting for events… Run <code className="text-violet-400">agentops dev</code> and use your agents.
          </div>
        )}
        {events.map((e) => (
          <EventRow key={e.id} event={e} />
        ))}
      </div>
    </div>
  );
}
