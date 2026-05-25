"use client";

import { useState } from "react";
import { api } from "@/lib/api";

export function RestoreButton({ sessionId, timestamp }: { sessionId: string; timestamp: string }) {
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);

  async function confirm() {
    setLoading(true);
    try {
      const snaps = await api.snapshots();
      const target = snaps.find((s) => s.timestamp <= timestamp) || snaps[0];
      if (target) {
        await api.restore(sessionId, target.id);
      }
      setOpen(false);
    } finally {
      setLoading(false);
    }
  }

  return (
    <>
      <button
        onClick={() => setOpen(true)}
        className="text-xs px-2 py-1 rounded bg-zinc-800 hover:bg-violet-500/20 text-violet-300"
      >
        ↩ Restore to here
      </button>
      {open && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
          <div className="card p-6 max-w-md">
            <h3 className="font-semibold mb-2">Restore working directory?</h3>
            <p className="text-sm text-zinc-400 mb-4">
              This will reset your working directory to {new Date(timestamp).toLocaleString()}. Continue?
            </p>
            <div className="flex gap-2 justify-end">
              <button onClick={() => setOpen(false)} className="px-3 py-1.5 text-sm text-zinc-400">
                Cancel
              </button>
              <button
                onClick={confirm}
                disabled={loading}
                className="px-3 py-1.5 text-sm rounded bg-violet-500 text-white"
              >
                {loading ? "Restoring…" : "Restore"}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
