# homebrew-agentops

Homebrew tap for [AgentOps](https://github.com/gog1withme/AgentOps).

## Setup

Create a public GitHub repository named `homebrew-agentops` under the `gog1withme` account. Goreleaser updates `Formula/agentops.rb` automatically on each release tag.

## Install

```bash
brew tap gog1withme/homebrew-agentops
brew install agentops
```

## Manual formula bootstrap

If publishing the first release before Goreleaser has pushed a formula, copy `Formula/agentops.rb` from this directory into the tap repository.
