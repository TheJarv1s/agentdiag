package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunScanAndExport(t *testing.T) {
	hermesHome := t.TempDir()
	ompHome := t.TempDir()
	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(hermesHome, "config.yaml"), []byte("providers:\n  x:\n    api_key: never-print-this\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(ompHome, "agent"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ompHome, "agent", "config.yml"), []byte("startup:\n  quiet: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	code := run([]string{"scan", "--project", project, "--hermes-home", hermesHome, "--omp-home", ompHome, "--format", "json"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
	if !strings.Contains(out.String(), `"name": "Hermes"`) || !strings.Contains(out.String(), `"name": "OMP"`) {
		t.Fatalf("output=%s", out.String())
	}
	if strings.Contains(out.String(), "never-print-this") {
		t.Fatal("secret leaked")
	}

	exportPath := filepath.Join(t.TempDir(), "report.md")
	out.Reset()
	errOut.Reset()
	code = run([]string{"export", "--project", project, "--hermes-home", hermesHome, "--omp-home", ompHome, "--format", "markdown", "--output", exportPath}, &out, &errOut)
	if code != 0 {
		t.Fatalf("export code=%d stderr=%s", code, errOut.String())
	}
	b, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "# AgentDiag report") || strings.Contains(string(b), "never-print-this") {
		t.Fatalf("bad export: %s", string(b))
	}
}

func TestRunVersion(t *testing.T) {
	var out, errOut bytes.Buffer
	if code := run([]string{"version"}, &out, &errOut); code != 0 {
		t.Fatalf("code=%d", code)
	}
	if !strings.Contains(out.String(), "0.1.1") {
		t.Fatalf("output=%s", out.String())
	}
}
