# AgentDiag

[![CI](https://github.com/TheJarv1s/agentdiag/actions/workflows/ci.yml/badge.svg)](https://github.com/TheJarv1s/agentdiag/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/github/license/TheJarv1s/agentdiag)](LICENSE)

English | [Русский](README.ru.md)

![AgentDiag diagnostics map](assets/agentdiag-social.png)

**AgentDiag** is a read-only diagnostics CLI for AI-agent environments. It detects **Hermes Agent** and **Oh My Pi (OMP)**, inventories skills, plugins and MCP configuration, identifies common discovery/configuration/security problems, and exports sanitized support reports.

**Current version:** 0.1.1 · **Platforms:** Windows, macOS, Linux · **Go:** 1.22+

> AgentDiag never starts agent binaries or executes plugins, hooks, modules, or skills.

## Why

Agent environments collect skills, plugins, MCP servers, project overrides, and provider configuration across several directories. When something stops loading, the question is not only whether a file exists: will the agent discover it, is it enabled, and is the configuration structurally valid?

AgentDiag answers those questions without changing the environment it inspects.

## What it checks

- Hermes and OMP detection on Windows, macOS, and Linux.
- User, project, and managed skills where applicable.
- Hermes plugins and OMP runtime plugin state.
- Native MCP server definitions and OMP MCP JSON validity.
- Duplicate skills and OMP nested layouts that native discovery ignores.
- Hermes plugin enablement and selected security-sensitive settings.
- Credential-like YAML literals by **key path only**—never the value.
- Unix secret-file permissions, terminal output, and JSON/Markdown exports.

## Install

```bash
go install github.com/TheJarv1s/agentdiag/cmd/agentdiag@latest
```

Or build from source:

```bash
go build -o agentdiag ./cmd/agentdiag
```

Windows PowerShell:

```powershell
go build -o agentdiag.exe ./cmd/agentdiag
```

## Quick start

```bash
agentdiag scan
agentdiag doctor
agentdiag security
agentdiag export --format markdown --output agentdiag-report.md
```

The current directory is the project by default, so project-local `.hermes` and `.omp` state is included.

## Real safe demo

The checked-in [safe demo fixture](examples/safe-demo/README.md) contains no credentials or personal paths. Run it from the repository root:

```bash
go run ./cmd/agentdiag scan \
  --project ./examples/safe-demo/project \
  --hermes-home ./examples/safe-demo/hermes \
  --omp-home ./examples/safe-demo/omp
```

![Terminal output from the reproducible safe demo](assets/agentdiag-terminal-demo.svg)

The screenshot is generated from that command; its paths are replaced with `<demo>` only for publication. See the [sanitized captured output](examples/safe-demo/terminal-output.txt).

## CLI

```text
agentdiag [scan] [flags]
agentdiag doctor [flags]
agentdiag security [flags]
agentdiag export --format markdown|json --output FILE [flags]
agentdiag version
```

Common overrides:

```text
--project PATH
--hermes-home PATH
--omp-home PATH
--omp-agent-dir PATH
--format terminal|json|markdown
```

## Diagnostic confidence and scope

AgentDiag distinguishes confirmed structural findings from heuristics. Heuristics appear as `POSSIBLE` (for example, `[WARNING/POSSIBLE]`). The v0.1.1 parser intentionally fails open for `SKILL.md` frontmatter: unsupported YAML syntax is not labeled invalid merely because the lightweight parser cannot represent it.

For Hermes, AgentDiag understands `plugins.enabled`, `plugins.disabled`, `mcp_servers`, selected skill-safety settings, plugin manifests, recursive skill trees, active `_org` mirrors, and configured `skills.external_dirs`.

For OMP, AgentDiag inspects the user layer (normally `~/.omp/agent`) and the current project `.omp` layer, including skills, managed skills, MCP files, and runtime plugin state. `PI_CODING_AGENT_DIR` and existing XDG OMP plugin data are recognized.

Current limitation: v0.1.1 inventories discoverable Hermes skill files. It does not emulate every trusted-project quarantine or platform/environment filter, so a listed skill is not a promise that it is active in the current session.

## Privacy and security

AgentDiag is intentionally conservative:

- read-only filesystem access;
- no shelling out to agent binaries;
- no plugin/module execution;
- no `.env` value parsing;
- no credential values in generated reports;
- literal-secret detection reports only a key path such as `providers.demo.api_key`.

Review a report before publishing it: paths and plugin/skill names can themselves reveal a project or username. See [SECURITY.md](SECURITY.md) for the project security policy and guarantees.

## Development and releases

```bash
go test ./...
go vet ./...
go build -trimpath ./cmd/agentdiag
```

Release builds are emitted into `dist/` by `make release`. The v0.1.1 release notes, asset plan, GitHub metadata, and promotion drafts are in [docs/launch](docs/launch/).

## Roadmap

- **v0.2:** Claude Code, Codex, and OpenCode adapters.
- **v0.2:** context/token overhead estimates and deeper OMP marketplace checks.
- **v0.3:** a machine-readable check registry, GitHub Actions mode, and an optional `--fix-plan` that suggests changes without applying them.

## Contributing

Contributions are welcome. Start with [CONTRIBUTING.md](CONTRIBUTING.md); adapters must remain read-only and must not execute third-party code.

## License

MIT. See [LICENSE](LICENSE).
