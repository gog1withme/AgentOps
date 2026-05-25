"use client";

import { useEffect, useState } from "react";
import { api } from "@/lib/api";
import type { BudgetResponse } from "@/lib/types";

export function BudgetBanner() {
  const [data, setData] = useState<BudgetResponse | null>(null);
  const [dismissed, setDismissed] = useState(false);

  useEffect(() => {
    api.budget().then(setData).catch(() => {});
    const t = setInterval(() => api.budget().then(setData).catch(() => {}), 5000);
    return () => clearInterval(t);
  }, []);

  if (!data || dismissed) return null;
  const { budget, spend } = data;
  if (budget.cost_limit_usd <= 0 && budget.tool_call_limit <= 0) return null;

  const costPct = budget.cost_limit_usd > 0 ? (spend.cost_usd / budget.cost_limit_usd) * 100 : 0;
  const toolPct = budget.tool_call_limit > 0 ? (spend.tool_calls / budget.tool_call_limit) * 100 : 0;
  const pct = Math.max(costPct, toolPct);
  if (pct < 80) return null;

  const isCritical = pct >= 100;
  return (
    <div
      className={`px-6 py-3 flex items-center justify-between text-sm ${
        isCritical ? "bg-red-500/20 text-red-200" : "bg-amber-500/20 text-amber-200"
      }`}
    >
      <span>
        ⚠ Budget ${spend.cost_usd.toFixed(2)} / ${budget.cost_limit_usd.toFixed(2)} ({pct.toFixed(0)}%) ·{" "}
        {spend.tool_calls} / {budget.tool_call_limit} tools
      </span>
      <button onClick={() => setDismissed(true)} className="text-zinc-400 hover:text-zinc-100">
        Dismiss
      </button>
    </div>
  );
}
