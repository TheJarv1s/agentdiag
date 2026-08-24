package omp

import (
	"os"
	"path/filepath"
	"testing"
)

func put(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestScanOMPInventoriesNativeSurfaces(t *testing.T) {
	home := t.TempDir()
	agent := filepath.Join(home, "agent")
	project := t.TempDir()
	put(t, filepath.Join(agent, "config.yml"), "providers:\n  demo:\n    api_key: literal-omp-secret\n")
	put(t, filepath.Join(agent, "skills", "good", "SKILL.md"), "---\nname: good\ndescription: Good skill\n---\n")
	put(t, filepath.Join(agent, "skills", "group", "nested", "SKILL.md"), "---\nname: nested\ndescription: Nested\n---\n")
	put(t, filepath.Join(agent, "managed-skills", "learned", "SKILL.md"), "---\nname: learned\ndescription: Learned\n---\n")
	put(t, filepath.Join(project, ".omp", "skills", "project-skill", "SKILL.md"), "---\nname: project-skill\n---\n")
	put(t, filepath.Join(agent, "mcp.json"), `{"mcpServers":{"files":{"command":"npx"}}}`)
	put(t, filepath.Join(project, ".omp", "mcp.json"), `{broken`)
	put(t, filepath.Join(home, "plugins", "omp-plugins.lock.json"), `{"plugins":{"tool-a":{"enabled":true},"tool-b":{"enabled":false}}}`)

	r := Scan(Options{Home: home, AgentDir: agent, Project: project})
	if !r.Detected {
		t.Fatal("expected detected")
	}
	if len(r.Skills) != 3 {
		t.Fatalf("skills=%d %+v", len(r.Skills), r.Skills)
	}
	if len(r.Plugins) != 2 {
		t.Fatalf("plugins=%d %+v", len(r.Plugins), r.Plugins)
	}
	if r.MCPServers != 1 {
		t.Fatalf("mcp=%d", r.MCPServers)
	}
	ids := map[string]bool{}
	for _, f := range r.Findings {
		ids[f.ID] = true
		if f.ID == "config.literal_secret" && f.Message == "literal-omp-secret" {
			t.Fatal("secret leaked")
		}
	}
	for _, id := range []string{"omp.skill_nested_ignored", "omp.skill_missing_description", "omp.mcp_invalid", "config.literal_secret"} {
		if !ids[id] {
			t.Fatalf("missing %s; findings=%+v", id, r.Findings)
		}
	}
}

func TestScanOMPDetectsXDGOnlyPluginState(t *testing.T) {
	xdg := t.TempDir()
	put(t, filepath.Join(xdg, "omp", "plugins", "omp-plugins.lock.json"), `{"plugins":{"xdg-tool":{"enabled":true}}}`)
	t.Setenv("XDG_DATA_HOME", xdg)
	t.Setenv("PI_CODING_AGENT_DIR", filepath.Join(t.TempDir(), "missing-agent"))
	r := Scan(Options{})
	if !r.Detected {
		t.Fatal("expected XDG-only OMP plugin state to trigger detection")
	}
	found := false
	for _, p := range r.Plugins {
		if p.Name == "xdg-tool" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected xdg-tool, got %+v", r.Plugins)
	}
}
