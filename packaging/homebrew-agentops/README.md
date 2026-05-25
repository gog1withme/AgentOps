# Homebrew tap setup

Goreleaser publishes the `agentops` formula to `gog1withme/homebrew-agentops` on each release tag.

## One-time setup

1. Create a public GitHub repository named `homebrew-agentops` under your account.
2. Push the bootstrap formula:

```bash
cd packaging/homebrew-agentops
git init
git checkout -b main
git add Formula README.md
git commit -m "Initial Homebrew tap for AgentOps"
git remote add origin https://github.com/gog1withme/homebrew-agentops.git
git push -u origin main
```

3. Ensure the GitHub Actions release workflow token can push to the tap repo (default `GITHUB_TOKEN` works when both repos are under the same owner).

## Install

```bash
brew tap gog1withme/homebrew-agentops
brew install agentops
```

After the first successful release, Goreleaser replaces `Formula/agentops.rb` automatically.
