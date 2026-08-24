package fsx

import (
	"os"
	"path/filepath"
	"reflect"
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
