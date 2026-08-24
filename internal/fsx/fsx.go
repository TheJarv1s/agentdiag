package fsx

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type SkillMeta struct {
	Name        string
	Description string
}

var frontmatterFencePattern = regexp.MustCompile(`^---\s*$`)

func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func IsDir(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.IsDir()
}

func LoadYAML(path string) (map[string]any, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseYAML(b)
}

type yamlLine struct {
	indent int
	text   string
}

func ParseYAML(b []byte) (map[string]any, error) {
	raw := strings.ReplaceAll(string(b), "\r\n", "\n")
	var lines []yamlLine
	for n, line := range strings.Split(raw, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") || trimmed == "---" || trimmed == "..." {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		if strings.Contains(line[:indent], "\t") {
			return nil, fmt.Errorf("line %d: tabs are not supported in indentation", n+1)
		}
		text := stripYAMLComment(strings.TrimSpace(line))
		if text != "" {
			lines = append(lines, yamlLine{indent: indent, text: text})
		}
	}
	if len(lines) == 0 {
		return map[string]any{}, nil
	}
	if strings.HasPrefix(lines[0].text, "- ") {
		return nil, errors.New("top-level YAML sequence is not supported")
	}
	v, next, err := parseYAMLMap(lines, 0, lines[0].indent)
	if err != nil {
		return nil, err
	}
	if next != len(lines) {
		return nil, fmt.Errorf("unexpected YAML content near %q", lines[next].text)
	}
	return v, nil
}

func parseYAMLMap(lines []yamlLine, i, indent int) (map[string]any, int, error) {
	out := map[string]any{}
	for i < len(lines) {
		line := lines[i]
		if line.indent < indent {
			break
		}
		if line.indent > indent {
			return nil, i, fmt.Errorf("unexpected indentation near %q", line.text)
		}
		if strings.HasPrefix(line.text, "- ") {
			break
		}
		key, value, ok := strings.Cut(line.text, ":")
		if !ok {
			return nil, i, fmt.Errorf("expected key: value near %q", line.text)
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return nil, i, errors.New("empty YAML key")
		}
		value = strings.TrimSpace(value)
		if value != "" {
			out[key] = parseYAMLScalar(value)
			i++
			continue
		}
		i++
		if i >= len(lines) || lines[i].indent <= indent {
			out[key] = map[string]any{}
			continue
		}
		childIndent := lines[i].indent
		if strings.HasPrefix(lines[i].text, "- ") {
			list, ni, err := parseYAMLList(lines, i, childIndent)
			if err != nil {
				return nil, i, err
			}
			out[key] = list
			i = ni
		} else {
			child, ni, err := parseYAMLMap(lines, i, childIndent)
			if err != nil {
				return nil, i, err
			}
			out[key] = child
			i = ni
		}
	}
	return out, i, nil
}

func parseYAMLList(lines []yamlLine, i, indent int) ([]any, int, error) {
	var out []any
	for i < len(lines) {
		line := lines[i]
		if line.indent < indent {
			break
		}
		if line.indent != indent || !strings.HasPrefix(line.text, "- ") {
			break
		}
		item := strings.TrimSpace(strings.TrimPrefix(line.text, "- "))
		if item == "" {
			out = append(out, nil)
			i++
			continue
		}
		if key, value, ok := strings.Cut(item, ":"); ok && strings.TrimSpace(key) != "" {
			m := map[string]any{strings.TrimSpace(key): parseYAMLScalar(strings.TrimSpace(value))}
			i++
			if i < len(lines) && lines[i].indent > indent && !strings.HasPrefix(lines[i].text, "- ") {
				child, ni, err := parseYAMLMap(lines, i, lines[i].indent)
				if err != nil {
					return nil, i, err
				}
				for k, v := range child {
					m[k] = v
				}
				i = ni
			}
			out = append(out, m)
			continue
		}
		out = append(out, parseYAMLScalar(item))
		i++
	}
	return out, i, nil
}

func parseYAMLScalar(s string) any {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		return s[1 : len(s)-1]
	}
	switch strings.ToLower(s) {
	case "true":
		return true
	case "false":
		return false
	case "null", "~":
		return nil
	}
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		inner := strings.TrimSpace(s[1 : len(s)-1])
		if inner == "" {
			return []any{}
		}
		parts := strings.Split(inner, ",")
		out := make([]any, 0, len(parts))
		for _, p := range parts {
			out = append(out, parseYAMLScalar(strings.TrimSpace(p)))
		}
		return out
	}
	return s
}

func stripYAMLComment(s string) string {
	var quote rune
	for i, r := range s {
		if r == '\'' || r == '"' {
			if quote == 0 {
				quote = r
			} else if quote == r {
				quote = 0
			}
		}
		if r == '#' && quote == 0 && (i == 0 || s[i-1] == ' ') {
			return strings.TrimSpace(s[:i])
		}
	}
	return strings.TrimSpace(s)
}

func LoadJSON(path string) (map[string]any, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, nil
}

func ParseSkill(path string) (SkillMeta, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return SkillMeta{}, err
	}
	text := strings.ReplaceAll(string(b), "\r\n", "\n")
	text = strings.TrimPrefix(text, "\ufeff")
	meta := SkillMeta{Name: filepath.Base(filepath.Dir(path))}
	if !strings.HasPrefix(text, "---") {
		return meta, nil
	}

	lines := strings.Split(text, "\n")
	if len(lines) == 0 || !frontmatterFencePattern.MatchString(lines[0]) {
		return meta, nil
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if frontmatterFencePattern.MatchString(lines[i]) {
			end = i
			break
		}
	}
	// Hermes fails open when frontmatter has no closing fence: the skill still
	// exists and falls back to its directory name rather than being rejected.
	if end < 0 {
		return meta, nil
	}

	front := strings.Join(lines[1:end], "\n")
	raw, parseErr := ParseYAML([]byte(front))
	if parseErr != nil {
		raw = parseFrontmatterFallback(front)
	}
	if name, ok := raw["name"].(string); ok && strings.TrimSpace(name) != "" {
		meta.Name = strings.TrimSpace(name)
	}
	if desc, ok := raw["description"].(string); ok {
		meta.Description = strings.TrimSpace(desc)
	}
	return meta, nil
}

// parseFrontmatterFallback mirrors Hermes' fail-open behavior for skill
// metadata. AgentDiag intentionally does not claim that a SKILL.md is invalid
// merely because its tiny config YAML parser cannot represent every valid YAML
// construct that PyYAML accepts. We only need top-level name/description here.
func parseFrontmatterFallback(front string) map[string]any {
	out := map[string]any{}
	lines := strings.Split(front, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		if indent != 0 {
			continue
		}
		key, value, ok := strings.Cut(strings.TrimSpace(line), ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			continue
		}
		if value == ">" || value == "|" {
			style := value
			var parts []string
			for j := i + 1; j < len(lines); j++ {
				next := lines[j]
				if strings.TrimSpace(next) == "" {
					parts = append(parts, "")
					i = j
					continue
				}
				nextIndent := len(next) - len(strings.TrimLeft(next, " "))
				if nextIndent <= indent {
					break
				}
				parts = append(parts, strings.TrimSpace(next))
				i = j
			}
			if style == ">" {
				out[key] = strings.TrimSpace(strings.Join(parts, " "))
			} else {
				out[key] = strings.TrimSpace(strings.Join(parts, "\n"))
			}
			continue
		}
		out[key] = parseYAMLScalar(value)
	}
	return out
}

var hermesExcludedSkillDirs = map[string]bool{
	".git": true, ".github": true, ".hub": true, ".archive": true,
	".venv": true, "venv": true, "node_modules": true, "site-packages": true,
	"__pycache__": true, ".tox": true, ".nox": true, ".pytest_cache": true,
	".mypy_cache": true, ".ruff_cache": true,
}

var hermesSkillSupportDirs = map[string]bool{
	"references": true, "templates": true, "assets": true, "scripts": true,
}

// FindHermesSkillFiles follows the active Hermes skill index pruning rules:
// VCS/dependency/cache trees are skipped, support packages below a real skill
// root are not indexed as standalone skills, and only the active _org mirror
// is traversed.
func FindHermesSkillFiles(root string) ([]string, error) {
	if !IsDir(root) {
		return nil, nil
	}
	root = filepath.Clean(root)
	orgRoot := filepath.Join(root, "_org")
	activeOrg := ""
	if b, err := os.ReadFile(filepath.Join(orgRoot, ".active_org")); err == nil {
		activeOrg = strings.TrimSpace(string(b))
	}

	var out []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path == root {
				return nil
			}
			name := d.Name()
			if hermesExcludedSkillDirs[name] {
				return filepath.SkipDir
			}
			if path == orgRoot && activeOrg == "" {
				return filepath.SkipDir
			}
			if filepath.Dir(path) == orgRoot && name != activeOrg {
				return filepath.SkipDir
			}
			if hermesSkillSupportDirs[name] && Exists(filepath.Join(filepath.Dir(path), "SKILL.md")) {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.EqualFold(d.Name(), "SKILL.md") {
			out = append(out, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(out)
	return out, nil
}

func FindSkillFiles(root string, recursive bool) ([]string, error) {
	if !IsDir(root) {
		return nil, nil
	}
	var out []string
	if !recursive {
		entries, err := os.ReadDir(root)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			p := filepath.Join(root, e.Name(), "SKILL.md")
			if Exists(p) {
				out = append(out, p)
			}
		}
	} else {
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.EqualFold(d.Name(), "SKILL.md") {
				out = append(out, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Strings(out)
	return out, nil
}

func isSecretKeyName(key string) bool {
	n := strings.ToLower(strings.TrimSpace(key))
	n = strings.NewReplacer("-", "_", " ", "_").Replace(n)
	suffixes := []string{
		"api_key", "token", "secret", "password", "credential", "credentials", "private_key",
	}
	for _, suffix := range suffixes {
		if n == suffix || strings.HasSuffix(n, "_"+suffix) {
			return true
		}
	}
	return false
}

func SecretKeyPaths(v any) []string {
	var out []string
	var walk func(any, string)
	walk = func(cur any, prefix string) {
		switch x := cur.(type) {
		case map[string]any:
			keys := make([]string, 0, len(x))
			for k := range x {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				val := x[k]
				path := k
				if prefix != "" {
					path = prefix + "." + k
				}
				if isSecretKeyName(k) && isLiteralSecret(val) {
					out = append(out, path)
				}
				walk(val, path)
			}
		case []any:
			for i, item := range x {
				walk(item, fmt.Sprintf("%s[%d]", prefix, i))
			}
		}
	}
	walk(v, "")
	sort.Strings(out)
	return out
}

func isLiteralSecret(v any) bool {
	s, ok := v.(string)
	if !ok {
		return false
	}
	s = strings.TrimSpace(s)
	if s == "" || strings.HasPrefix(s, "${") || strings.HasPrefix(s, "$env:") || strings.HasPrefix(s, "$ENV:") {
		return false
	}
	return true
}

func CountMCPServersJSON(path string) (int, error) {
	obj, err := LoadJSON(path)
	if err != nil {
		return 0, err
	}
	for _, key := range []string{"mcpServers", "servers"} {
		if m, ok := obj[key].(map[string]any); ok {
			return len(m), nil
		}
	}
	return 0, nil
}

func StringSlice(v any) []string {
	var out []string
	switch x := v.(type) {
	case []any:
		for _, item := range x {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
	case []string:
		out = append(out, x...)
	}
	return out
}

func MapAt(m map[string]any, keys ...string) map[string]any {
	cur := m
	for _, k := range keys {
		v, ok := cur[k]
		if !ok {
			return nil
		}
		next, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		cur = next
	}
	return cur
}

func ValueAt(m map[string]any, keys ...string) any {
	if len(keys) == 0 {
		return m
	}
	cur := m
	for i, k := range keys {
		v, ok := cur[k]
		if !ok {
			return nil
		}
		if i == len(keys)-1 {
			return v
		}
		next, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		cur = next
	}
	return nil
}

func WorldReadable(path string) bool {
	st, err := os.Stat(path)
	if err != nil {
		return false
	}
	return st.Mode().Perm()&0o077 != 0
}
