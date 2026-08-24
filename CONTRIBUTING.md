# Contributing

AgentDiag keeps adapters isolated so agent-specific behavior can evolve independently.

## Setup

```bash
go test ./...
go vet ./...
```

## Adding a check

1. Reproduce the behavior with a temp-directory fixture test.
2. Add the finding in the relevant adapter.
3. Never include secret values in `Message`, `Path` or `Remediation`.
4. Prefer a stable finding ID (`agent.area_problem`) because future CI users may depend on it.
5. Run the full test suite and vet.

## Adding an agent

Create a new adapter under `internal/adapters/<agent>` that returns `model.AgentReport`, then register it in `internal/engine`.

Adapters must be read-only and must not execute third-party code.
