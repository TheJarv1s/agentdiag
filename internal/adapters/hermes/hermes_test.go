package hermes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/TheJarv1s/agentdiag/internal/model"
)

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func hasFinding(rID string, findings []string) bool {
	for _, id := range findings {
		if id == rID {
			return true
		}
	}
	return false
}

func TestScanHermesInventoriesAndDiagnoses(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	writeFile(t, filepath.Join(home, "config.yaml"), `plugins:
  enabled:
    - good-plugin
  disabled:
    - conflicted
  entries: {}
skills:
  guard_agent_created: false
  write_approval: false
mcp_servers:
  filesystem:
    command: npx
  browser:
    command: node
providers:
  demo:
    api_key: literal-secret-should-never-be-reported
`)
	writeFile(t, filepath.Join(home, "skills", "one", "SKILL.md"), "---\nname: duplicate\ndescription: One\n---\n")
	writeFile(t, filepath.Join(home, "skills", "group", "two", "SKILL.md"), "---\nname: duplicate\ndescription: Two\n---\n")
	writeFile(t, filepath.Join(home, "plugins", "good-plugin", "plugin.yaml"), "name: good-plugin\nversion: 1.0\n")
	writeFile(t, filepath.Join(home, "plugins", "unused-plugin", "plugin.yaml"), "name: unused-plugin\n")
	writeFile(t, filepath.Join(project, ".hermes", "plugins", "project-tool", "plugin.yaml"), "name: project-tool\n")
	t.Setenv("HERMES_ENABLE_PROJECT_PLUGINS", "true")

	r := Scan(Options{Home: home, Project: project})
	if !r.Detected {
		t.Fatal("expected detected")
	}
	if len(r.Skills) != 2 {
		t.Fatalf("skills=%d", len(r.Skills))
	}
	if len(r.Plugins) != 2 {
		t.Fatalf("plugins=%d", len(r.Plugins))
	}
	if r.MCPServers != 2 {
		t.Fatalf("mcp=%d", r.MCPServers)
	}

	var ids []string
	for _, f := range r.Findings {
		ids = append(ids, f.ID)
		if f.ID == "config.literal_secret" && f.Message == "literal-secret-should-never-be-reported" {
			t.Fatal("secret value leaked in finding")
		}
	}
	for _, want := range []string{"hermes.skill_duplicate", "hermes.plugin_not_enabled", "hermes.project_plugins_enabled", "config.literal_secret", "hermes.skill_guard_disabled"} {
		if !hasFinding(want, ids) {
			t.Fatalf("missing finding %s; got %v", want, ids)
		}
	}
}

func TestPermissionCheckPlatformGate(t *testing.T) {
	if permissionCheckEnabled("windows") {
		t.Fatal("Windows permission bits are not meaningful for this check")
	}
	if !permissionCheckEnabled("linux") || !permissionCheckEnabled("darwin") {
		t.Fatal("expected Unix permission checks")
	}
}

func TestScanHermesIgnoresExcludedAndSupportSkillCopies(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, "config.yaml"), "model:\n  max_tokens: 32768\n")
	writeFile(t, filepath.Join(home, "skills", "media", "youtube-content", "SKILL.md"), "---\nname: youtube-content\n---\n")
	writeFile(t, filepath.Join(home, "skills", ".archive", "youtube-content", "SKILL.md"), "---\nname: youtube-content\n---\n")
	writeFile(t, filepath.Join(home, "skills", "media", "youtube-content", "references", "old-package", "SKILL.md"), "---\nname: youtube-content\n---\n")

	r := Scan(Options{Home: home})
	if len(r.Skills) != 1 {
		t.Fatalf("skills=%d %+v", len(r.Skills), r.Skills)
	}
	for _, f := range r.Findings {
		if f.ID == "config.literal_secret" || f.ID == "hermes.skill_duplicate" {
			t.Fatalf("unexpected false positive: %+v", f)
		}
	}
}

func TestScanHermesIncludesExternalSkillDirs(t *testing.T) {
	home := t.TempDir()
	external := filepath.Join(home, "shared-skills")
	writeFile(t, filepath.Join(home, "config.yaml"), "skills:\n  external_dirs:\n    - shared-skills\n")
	writeFile(t, filepath.Join(home, "skills", "local", "SKILL.md"), "---\nname: local\n---\n")
	writeFile(t, filepath.Join(external, "external", "SKILL.md"), "---\nname: external\n---\n")

	r := Scan(Options{Home: home})
	if len(r.Skills) != 2 {
		t.Fatalf("skills=%d %+v", len(r.Skills), r.Skills)
	}
	if r.Skills[0].Source != "local" || r.Skills[1].Source != "external" {
		t.Fatalf("expected source labels, got %+v", r.Skills)
	}
}

func TestScanHermesDoesNotCallUnsupportedYAMLInvalid(t *testing.T) {
	home := t.TempDir()
	writeFile(t, filepath.Join(home, "config.yaml"), "notes: >\n  valid YAML block scalar\n  that the lightweight parser does not implement\n")

	r := Scan(Options{Home: home})
	var foundPossible bool
	for _, f := range r.Findings {
		if f.ID == "hermes.config_invalid" || (f.Severity == "error" && strings.Contains(f.ID, "config")) {
			t.Fatalf("unsupported YAML must not be reported as confirmed invalid: %+v", f)
		}
		if f.ID == "hermes.config_unparsed" && f.Confidence == model.ConfidencePossible {
			foundPossible = true
		}
	}
	if !foundPossible {
		t.Fatalf("expected conservative unparsed finding, got %+v", r.Findings)
	}
}
