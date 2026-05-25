"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import type { AlertRecord } from "@/lib/types";

export default function SecurityPage() {
  const [alerts, setAlerts] = useState<AlertRecord[]>([]);

  useEffect(() => {
    api.alerts().then(setAlerts).catch(() => {});
  }, []);

  return (
    <div>
      <h1 className="text-2xl font-semibold mb-6">Security Center</h1>
      <div className="space-y-3">
        {alerts.map((a) => (
          <div
            key={a.id}
            className={`card p-4 border-l-4 ${
              a.severity === "critical" ? "border-l-red-500" : "border-l-amber-500"
            }`}
          >
            <div className="flex items-center gap-2 text-sm">
              <span className="uppercase text-xs text-zinc-500">{a.rule}</span>
              <span className={`text-xs ${a.severity === "critical" ? "text-red-400" : "text-amber-400"}`}>
                {a.severity}
              </span>
            </div>
            <p className="mt-1 text-sm">{a.message}</p>
            <p className="text-xs text-zinc-500 mt-2">{new Date(a.timestamp).toLocaleString()}</p>
          </div>
        ))}
        {alerts.length === 0 && (
          <div className="card p-8 text-center text-zinc-500">No security alerts.</div>
        )}
      </div>
    </div>
  );
}
