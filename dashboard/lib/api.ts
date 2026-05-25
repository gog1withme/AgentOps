const API_BASE =
  typeof window !== "undefined"
    ? `${window.location.protocol}//${window.location.host}`
    : process.env.NEXT_PUBLIC_API_URL || "http://127.0.0.1:4318";

async function fetchJSON<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...init,
    headers: { "Content-Type": "application/json", ...init?.headers },
  });
  if (!res.ok) {
    throw new Error(`API ${path}: ${res.status}`);
  }
  return res.json();
}

async function fetchJSONArray<T>(path: string, init?: RequestInit): Promise<T[]> {
  const data = await fetchJSON<T[] | null>(path, init);
  return Array.isArray(data) ? data : [];
}

export const api = {
  events: (limit = 100) => fetchJSONArray<import("./types").Event>(`/api/events?limit=${limit}`),
  sessions: () => fetchJSONArray<import("./types").SessionSummary>("/api/sessions"),
  replay: (id: string) => fetchJSONArray<import("./types").Event>(`/api/sessions/${id}/replay`),
  cost: (groupBy = "model") => fetchJSONArray<Record<string, unknown>>(`/api/cost?group_by=${groupBy}`),
  efficiency: () => fetchJSONArray<Record<string, unknown>>("/api/cost/efficiency"),
  budget: () => fetchJSON<import("./types").BudgetResponse>("/api/budget"),
  blame: (file: string) => fetchJSONArray<import("./types").Attribution>(`/api/blame/${encodeURIComponent(file)}`),
  mcp: () => fetchJSONArray<import("./types").MCPServer>("/api/mcp"),
  prompts: () => fetchJSONArray<import("./types").Prompt>("/api/prompts"),
  alerts: () => fetchJSONArray<import("./types").AlertRecord>("/api/alerts"),
  snapshots: () => fetchJSONArray<import("./types").Snapshot>("/api/snapshots"),
  restore: (sessionId: string, snapshotId: string) =>
    fetchJSON<{ ok: boolean }>("/api/restore", {
      method: "POST",
      body: JSON.stringify({ session_id: sessionId, snapshot_id: snapshotId }),
    }),
  streamURL: () => `${API_BASE}/api/stream`,
};

export { API_BASE };
