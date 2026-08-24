package hermes

import (
	"os"
	"path/filepath"
	"testing"
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
