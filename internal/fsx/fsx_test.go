package fsx

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadYAMLAndSecretKeyPaths(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := "provider:\n  api_key: sk-test-secret\n  model: demo\nauxiliary:\n  token: ${TOKEN}\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadYAML(path)
	if err != nil {
		t.Fatalf("LoadYAML: %v", err)
	}
	got := SecretKeyPaths(cfg)
	want := []string{"provider.api_key"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestParseSkillFrontmatter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(path, []byte("---\nname: sample\ndescription: Useful skill\n---\n# Body\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	meta, err := ParseSkill(path)
	if err != nil {
		t.Fatal(err)
	}
	if meta.Name != "sample" || meta.Description != "Useful skill" {
		t.Fatalf("unexpected meta: %+v", meta)
	}
}

func TestSecretKeyPathsDoesNotTreatTokenBudgetsAsCredentials(t *testing.T) {
	cfg := map[string]any{
		"model": map[string]any{
			"max_tokens":        "32768",
			"max_output_tokens": "8192",
			"token_budget":      "12000",
			"access_token":      "literal-access-token",
		},
	}
	got := SecretKeyPaths(cfg)
	want := []string{"model.access_token"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestParseSkillFallsBackForHermesCompatibleFrontmatter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "fallback-skill", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "\ufeff---\nname: fallback-skill\ndescription: >\n  Multi-line description that our minimal YAML parser\n  does not need to fully validate.\nmetadata:\n  nested:\n    - key: value\n      extra: true\n---\n# Body\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	meta, err := ParseSkill(path)
	if err != nil {
		t.Fatalf("ParseSkill should mirror Hermes fallback behavior, got error: %v", err)
	}
	if meta.Name != "fallback-skill" {
		t.Fatalf("name=%q", meta.Name)
	}
	if meta.Description == "" {
		t.Fatal("expected description from folded block scalar")
	}
}

func TestFindHermesSkillFilesMatchesDiscoveryPruning(t *testing.T) {
	root := t.TempDir()
	writeSkill := func(rel string) {
		t.Helper()
		p := filepath.Join(root, filepath.FromSlash(rel), "SKILL.md")
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("---\nname: "+filepath.Base(filepath.Dir(p))+"\n---\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	writeSkill("category/active")
	writeSkill(".archive/old-copy")
	writeSkill("node_modules/vendor-copy")
	writeSkill("parent")
	writeSkill("parent/references/archived-skill")
	writeSkill("scripts/legitimate-category")

	files, err := FindHermesSkillFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	var rel []string
	for _, p := range files {
		r, err := filepath.Rel(root, p)
		if err != nil {
			t.Fatal(err)
		}
		rel = append(rel, filepath.ToSlash(r))
	}
	want := []string{
		"category/active/SKILL.md",
		"parent/SKILL.md",
		"scripts/legitimate-category/SKILL.md",
	}
	if !reflect.DeepEqual(rel, want) {
		t.Fatalf("got %v want %v", rel, want)
	}
}

func TestFindHermesSkillFilesOnlyScansActiveOrgMirror(t *testing.T) {
	root := t.TempDir()
	orgRoot := filepath.Join(root, "_org")
	if err := os.MkdirAll(orgRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(orgRoot, ".active_org"), []byte("org-a\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"_org/org-a/live", "_org/org-b/stale"} {
		p := filepath.Join(root, filepath.FromSlash(rel), "SKILL.md")
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("---\nname: org-skill\n---\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	files, err := FindHermesSkillFiles(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || !strings.Contains(filepath.ToSlash(files[0]), "/_org/org-a/live/SKILL.md") {
		t.Fatalf("unexpected files: %v", files)
	}
}
