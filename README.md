# AgentDiag

**AgentDiag** is a read-only diagnostics CLI for AI-agent environments.

It detects **Hermes Agent** and **Oh My Pi (OMP)**, inventories skills/plugins/MCP configuration, catches common discovery/config/security problems, and can export a **sanitized support report** that is safe to attach to an issue.

> v0.1 focuses on Hermes + OMP. The core is adapter-based so more agents can be added without rewriting the CLI.

## Why

AI-agent setups accumulate skills, plugins, MCP servers, project overrides and provider configuration across several directories. When something silently stops loading, the useful question is often not “is the file there?” but “will this agent actually discover it, is it enabled, and is the config valid?”

AgentDiag answers that without starting the agent or executing third-party plugin code.

## v0.1 features

- Detects Hermes and OMP on Windows, macOS and Linux.
- Inventories user/project/managed skills where applicable.
- Inventories Hermes plugins and OMP runtime plugin state.
- Counts native MCP server definitions and validates OMP MCP JSON.
- Detects duplicate skill names and OMP nested skill layouts that native discovery ignores.
- Checks Hermes plugin enabled/disabled state.
- Flags credential-like literals in YAML by **key path only**; values are never included in output.
- Checks selected security-sensitive settings and Unix secret-file permissions.
- Exports terminal, JSON or Markdown reports.
- Does **not** execute `hermes`, `omp`, plugins, hooks or skills.

## Quick start

### Install with Go

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

### Scan the current environment

```bash
./agentdiag
```

or explicitly:

```bash
./agentdiag scan
```

### Full diagnostic view

```bash
./agentdiag doctor
```

### Security findings only

```bash
./agentdiag security
```

### Export a sanitized report

```bash
./agentdiag export --format markdown --output agentdiag-report.md
```

```bash
./agentdiag export --format json --output agentdiag-report.json
```

## Example

```text
AgentDiag v0.1.0

Hermes  [DETECTED]
  Home:    C:\Users\you\.hermes
  Skills:  18
  Plugins: 5
  MCP:     4
  Findings: 2
    [WARNING] hermes.skill_duplicate: Duplicate Hermes skill name `review` shadows another skill.
    [INFO] hermes.plugin_not_enabled: Plugin `tool-slimmer` is installed but not enabled.

OMP  [DETECTED]
  Home:    C:\Users\you\.omp
  Skills:  27
  Plugins: 8
  MCP:     6
  Findings: 1
    [WARNING] omp.skill_nested_ignored: Nested OMP skill layout is outside native provider discovery depth and will be ignored.

Summary: 0 error(s), 2 warning(s), 1 info
```

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

The current directory is scanned as the project by default so project-local `.hermes` / `.omp` state can be diagnosed.

## What v0.1 knows about Hermes

AgentDiag inspects the Hermes home (normally `~/.hermes`):

```text
config.yaml
.env
skills/
plugins/
```

It understands `plugins.enabled`, `plugins.disabled`, `mcp_servers`, selected skill safety settings, plugin manifests and recursive Hermes skill trees.

## What v0.1 knows about OMP

AgentDiag inspects the OMP user layer (normally `~/.omp/agent`) and current project `.omp` layer, including:

```text
~/.omp/agent/config.yml
~/.omp/agent/skills/*/SKILL.md
~/.omp/agent/managed-skills/*/SKILL.md
~/.omp/agent/mcp.json
~/.omp/plugins/omp-plugins.lock.json
~/.omp/plugins/installed_plugins.json
<project>/.omp/skills/*/SKILL.md
<project>/.omp/mcp.json
<project>/.omp/plugins/...
```

`PI_CODING_AGENT_DIR` is honored for the active user agent directory. Existing XDG OMP plugin data is also recognized.

## Privacy and security

AgentDiag is intentionally conservative:

- read-only filesystem access;
- no shelling out to agent binaries;
- no plugin/module execution;
- no `.env` value parsing;
- no credential values in generated reports;
- literal-secret detection reports only a dotted key path such as `providers.demo.api_key`.

Please still review a report before publishing it: paths and plugin/skill names can themselves reveal project or username information.

## Development

```bash
go test ./...
go vet ./...
```

Release builds:

```bash
make release
```

Artifacts are written to `dist/`.

## Repository

Source: `github.com/TheJarv1s/agentdiag`

Go module: `github.com/TheJarv1s/agentdiag`

## Roadmap

- **v0.2:** Claude Code, Codex and OpenCode adapters.
- **v0.2:** context/token overhead estimates.
- **v0.2:** deeper OMP marketplace consistency checks (registry ↔ runtime lock ↔ node_modules).
- **v0.3:** machine-readable check registry and GitHub Actions mode.
- **v0.3:** optional `--fix-plan` that suggests changes without applying them.

## License

MIT. See [LICENSE](LICENSE).
