# AgentDiag v0.1.1 release notes

## English

### Highlights

- Improved Hermes skill discovery to mirror exclusion/support-directory pruning and the active `_org` mirror.
- Added configured `skills.external_dirs` to Hermes discovery.
- Made `SKILL.md` frontmatter parsing conservative and fail-open, avoiding false invalid-skill warnings for valid YAML constructs outside the lightweight parser.
- Added confirmed-versus-possible confidence to terminal, Markdown, and JSON findings.
- Reduced false secret findings for token-budget settings such as `model.max_tokens`.

### Compatibility and safety

AgentDiag remains read-only: it does not execute agents, plugins, hooks, or skills, and reports credential-like YAML findings by key path rather than value.

### Release assets

Six standalone binaries are prepared for Windows (amd64, arm64), Linux (amd64, arm64), and macOS (amd64, arm64), with a `SHA256SUMS.txt` manifest.

## Русский

### Главное

- Улучшен Hermes skill discovery: он учитывает exclusion/support-directory pruning и активный `_org` mirror.
- Настроенные `skills.external_dirs` добавлены в Hermes discovery.
- Парсинг `SKILL.md` frontmatter стал консервативным и fail-open: валидные YAML-конструкции вне возможностей лёгкого парсера больше не дают ложных предупреждений о невалидном skill.
- В terminal, Markdown и JSON findings добавлена уверенность: confirmed или possible.
- Уменьшены ложные secret findings для token-budget settings вроде `model.max_tokens`.

### Совместимость и безопасность

AgentDiag остаётся read-only: он не исполняет agents, plugins, hooks или skills и сообщает о credential-like YAML findings по пути ключа, а не по значению.

### Release assets

Подготовлены шесть standalone binaries для Windows (amd64, arm64), Linux (amd64, arm64) и macOS (amd64, arm64) с манифестом `SHA256SUMS.txt`.
