# Safe AgentDiag demo

This fixture is a deliberately synthetic, credential-free environment for a reproducible AgentDiag demonstration. It represents two common, non-destructive findings:

- a potentially shadowed Hermes skill name; and
- a nested OMP skill that native discovery ignores.

Run from the repository root:

```bash
go run ./cmd/agentdiag scan \
  --project ./examples/safe-demo/project \
  --hermes-home ./examples/safe-demo/hermes \
  --omp-home ./examples/safe-demo/omp
```

`terminal-output.txt` is a captured result from that command. Absolute fixture paths have been replaced with `<demo>` for publication.
