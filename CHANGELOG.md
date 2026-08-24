# Changelog

All notable changes to AgentDiag are documented here.

## [Unreleased]

## [0.1.1] - 2026-08-25

### Fixed
- Prevent credential scanning from flagging token-budget fields such as `model.max_tokens`, `max_output_tokens`, and `token_budget`.
- Mirror Hermes skill-tree pruning for VCS/dependency/cache directories, progressive-disclosure support directories, and the active `_org` mirror.
- Include configured Hermes `skills.external_dirs` in discovery.
- Make SKILL.md frontmatter parsing fail open like Hermes: BOMs, folded/literal descriptions, and YAML constructs outside AgentDiag's lightweight parser no longer produce false `skill_invalid` warnings.
- Stop claiming unsupported YAML is definitely invalid; config/plugin parse limitations are reported as `POSSIBLE` findings.
- Add finding confidence to terminal, Markdown, and JSON reports.
- Refresh CI actions and disable unnecessary Go dependency caching for the zero-dependency build.

## [0.1.0] - 2026-08-19

### Added
- Hermes Agent detection, skill/plugin/MCP inventory and diagnostic checks.
- OMP detection, authored/managed/project skill inventory, plugin runtime inventory and MCP validation.
- Duplicate/malformed skill checks and OMP ignored nested-layout detection.
- Sanitized credential-key detection without reporting credential values.
- `scan`, `doctor`, `security`, `export`, and `version` commands.
- Terminal, JSON and Markdown reports.
- Windows, macOS and Linux cross-build targets.
