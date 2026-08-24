package engine

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunAndSecurityFilter(t *testing.T) {
	hermesHome := t.TempDir()
	ompHome := t.TempDir()
	if err := os.WriteFile(filepath.Join(hermesHome, "config.yaml"), []byte("providers:\n  x:\n    api_key: literal\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(ompHome, "agent"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ompHome, "agent", "config.yml"), []byte("startup:\n  quiet: true\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	r := Run(Options{HermesHome: hermesHome, OMPHome: ompHome})
	if len(r.Agents) != 2 || !r.Agents[0].Detected || !r.Agents[1].Detected {
		t.Fatalf("unexpected agents: %+v", r.Agents)
	}
	s := Filter(r, ModeSecurity)
	for _, a := range s.Agents {
		for _, f := range a.Findings {
			if !f.Security {
				t.Fatalf("non-security finding: %+v", f)
			}
		}
	}
}
