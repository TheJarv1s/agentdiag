# Security Policy

AgentDiag reads local AI-agent configuration, so privacy and secret handling are part of the core security model.

## Reporting a vulnerability

Please do not publish credential leaks or proof-of-concept reports containing real secrets in a public issue. Use the repository owner's private security reporting channel once the project is published.

## Design guarantees for v0.1

- AgentDiag does not execute Hermes, OMP, plugins, skills or hooks.
- `.env` contents are not loaded into reports.
- Secret-like YAML values are detected by key name; only the dotted key path is emitted.
- Export files are created with user-only permissions where the OS honors Unix-style permission bits.
