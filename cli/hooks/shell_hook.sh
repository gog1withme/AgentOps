#!/usr/bin/env bash
# AgentOps shell hook — sourced by agentops init
agentops_record() {
  local exit_code=$?
  if [ -n "$AGENTOPS_ENABLED" ]; then
    curl -s -X POST "http://127.0.0.1:${AGENTOPS_PORT:-4318}/api/ingest/shell" \
      -H "Content-Type: application/json" \
      -d "{\"command\":\"$1\",\"exit_code\":$exit_code}" >/dev/null 2>&1 || true
  fi
}
