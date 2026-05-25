"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import clsx from "clsx";

const links = [
  { href: "/", label: "Live Feed", icon: "⚡" },
  { href: "/traces/", label: "Traces", icon: "🕐" },
  { href: "/cost/", label: "Cost", icon: "💰" },
  { href: "/blame/", label: "Blame", icon: "📝" },
  { href: "/mcp/", label: "MCP Health", icon: "🔌" },
  { href: "/prompts/", label: "Prompts", icon: "📋" },
  { href: "/context/", label: "Context", icon: "🌳" },
  { href: "/security/", label: "Security", icon: "🛡" },
];

export function Sidebar() {
  const pathname = usePathname();
  return (
    <aside className="w-[220px] shrink-0 border-r border-zinc-800 bg-zinc-950 p-4 flex flex-col gap-1">
      <div className="mb-6 px-2">
        <div className="text-lg font-semibold text-violet-400">AgentOps</div>
        <div className="text-xs text-zinc-500">local observability</div>
      </div>
      {links.map((l) => (
        <Link
          key={l.href}
          href={l.href}
          className={clsx(
            "flex items-center gap-2 rounded-lg px-3 py-2 text-sm transition-colors",
            pathname === l.href || pathname.startsWith(l.href.slice(0, -1))
              ? "bg-violet-500/10 text-violet-300"
              : "text-zinc-400 hover:bg-zinc-900 hover:text-zinc-100"
          )}
        >
          <span>{l.icon}</span>
          {l.label}
        </Link>
      ))}
    </aside>
  );
}
