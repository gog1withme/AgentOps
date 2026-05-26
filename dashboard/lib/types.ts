export type EventType =
  | "llm_call"
  | "tool_call"
  | "file_edit"
  | "shell_command"
  | "mcp_call"
  | "budget_alert"
  | "restore"
  | "alert";

export interface Event {
  id: string;
  session_id: string;
  run_id?: string;
  timestamp: string;
  source: string;
  type: EventType;
  model?: string;
  prompt_hash?: string;
  prompt_tokens?: number;
  output_tokens?: number;
  efficiency_score?: number;
  cost_usd?: number;
  duration_ms?: number;
  tool_name?: string;
  tool_input?: string;
  tool_output?: string;
  file_path?: string;
  file_diff?: string;
  shell_command?: string;
  shell_exit_code?: number;
  mcp_server?: string;
  mcp_latency_ms?: number;
  error?: string;
  metadata?: Record<string, string>;
}

export interface SessionSummary {
  session_id: string;
  first_event: string;
  last_event: string;
  event_count: number;
  total_cost: number;
  total_tokens: number;
}

export interface BudgetResponse {
  budget: {
    cost_limit_usd: number;
    tool_call_limit: number;
    llm_call_limit: number;
    action: string;
  };
  spend: {
    cost_usd: number;
    tool_calls: number;
    llm_calls: number;
  };
}

export interface Attribution {
  id: string;
  session_id: string;
  timestamp: string;
  file_path: string;
  agent: string;
  lines_added: number;
  lines_removed: number;
  diff: string;
}

export interface MCPServer {
  name: string;
  url?: string;
  total_calls: number;
  error_count: number;
  avg_latency_ms: number;
  p95_latency_ms: number;
}

export interface Prompt {
  hash: string;
  content: string;
  first_seen: string;
  session_count: number;
  avg_cost_usd: number;
  avg_prompt_tokens: number;
  avg_efficiency_score: number;
}

export interface AlertRecord {
  id: string;
  rule: string;
  severity: string;
  message: string;
  timestamp: string;
}

export interface Snapshot {
  id: string;
  session_id: string;
  timestamp: string;
  trigger: string;
}

export interface ContextFileStat {
  path: string;
  occurrences: number;
  total_tokens: number;
  avg_efficiency: number;
  consecutive_hits?: number;
}

export interface ContextAnalysis {
  session_id: string;
  total_prompt_tokens: number;
  llm_call_count: number;
  avg_efficiency: number;
  duplicate_files: ContextFileStat[];
  noisy_files: ContextFileStat[];
  callouts: string[];
}
