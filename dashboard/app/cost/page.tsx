"use client";

import { useEffect, useState } from "react";
import { Bar, BarChart, CartesianGrid, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import { api } from "@/lib/api";

export default function CostPage() {
  const [cost, setCost] = useState<Record<string, unknown>[]>([]);
  const [efficiency, setEfficiency] = useState<Record<string, unknown>[]>([]);

  useEffect(() => {
    api.cost("model").then(setCost).catch(() => {});
    api.efficiency().then(setEfficiency).catch(() => {});
  }, []);

  const costData = cost.map((c) => ({
    name: String(c.key || "unknown").slice(0, 12),
    cost: Number(c.cost_usd || 0),
  }));

  const effData = efficiency.map((e) => ({
    time: new Date(String(e.timestamp)).toLocaleTimeString(),
    score: Number(e.efficiency_score || 0),
  }));

  return (
    <div>
      <h1 className="text-2xl font-semibold mb-6">Cost & Efficiency</h1>
      <div className="grid gap-6 lg:grid-cols-2">
        <div className="card p-4 h-80">
          <h2 className="text-sm font-medium text-zinc-400 mb-4">Cost by model</h2>
          <ResponsiveContainer width="100%" height="90%">
            <BarChart data={costData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
              <XAxis dataKey="name" stroke="#71717a" fontSize={12} />
              <YAxis stroke="#71717a" fontSize={12} />
              <Tooltip contentStyle={{ background: "#18181b", border: "1px solid #3f3f46" }} />
              <Bar dataKey="cost" fill="#8b5cf6" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
        <div className="card p-4 h-80">
          <h2 className="text-sm font-medium text-zinc-400 mb-4">Efficiency trend</h2>
          <ResponsiveContainer width="100%" height="90%">
            <LineChart data={effData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
              <XAxis dataKey="time" stroke="#71717a" fontSize={12} />
              <YAxis domain={[0, 100]} stroke="#71717a" fontSize={12} />
              <Tooltip contentStyle={{ background: "#18181b", border: "1px solid #3f3f46" }} />
              <Line type="monotone" dataKey="score" stroke="#10b981" dot={false} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  );
}
