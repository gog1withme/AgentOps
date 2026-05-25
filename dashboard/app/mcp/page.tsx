"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import type { MCPServer } from "@/lib/types";

export default function MCPPage() {
  const [servers, setServers] = useState<MCPServer[]>([]);

  useEffect(() => {
    api.mcp().then(setServers).catch(() => {});
    const t = setInterval(() => api.mcp().then(setServers).catch(() => {}), 5000);
    return () => clearInterval(t);
  }, []);

  return (
    <div>
      <h1 className="text-2xl font-semibold mb-6">MCP Health</h1>
      <div className="card overflow-hidden">
        <table className="w-full text-sm">
          <thead className="border-b border-zinc-800 text-zinc-500">
            <tr>
              <th className="text-left p-4">Server</th>
              <th className="text-right p-4">Calls</th>
              <th className="text-right p-4">Errors</th>
              <th className="text-right p-4">Avg ms</th>
              <th className="text-right p-4">P95 ms</th>
              <th className="text-left p-4">Status</th>
            </tr>
          </thead>
          <tbody>
            {servers.map((s) => {
              const errRate = s.total_calls > 0 ? s.error_count / s.total_calls : 0;
              const slow = s.p95_latency_ms > 2000;
              const failing = errRate > 0.1;
              return (
                <tr key={s.name} className="border-b border-zinc-800/50">
                  <td className="p-4 font-medium">{s.name}</td>
                  <td className="p-4 text-right">{s.total_calls}</td>
                  <td className="p-4 text-right">{s.error_count}</td>
                  <td className="p-4 text-right">{s.avg_latency_ms.toFixed(0)}</td>
                  <td className="p-4 text-right">{s.p95_latency_ms.toFixed(0)}</td>
                  <td className="p-4">
                    {failing && <span className="text-red-400 text-xs mr-2">✗ error rate {(errRate * 100).toFixed(0)}%</span>}
                    {slow && <span className="text-amber-400 text-xs">⚠ p95 &gt; 2s</span>}
                    {!failing && !slow && <span className="text-emerald-400 text-xs">● healthy</span>}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
        {servers.length === 0 && (
          <div className="p-8 text-center text-zinc-500">No MCP servers recorded yet.</div>
        )}
      </div>
    </div>
  );
}
