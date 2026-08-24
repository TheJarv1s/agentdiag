package report

import (
	"strings"
	"testing"
	"time"

	"github.com/TheJarv1s/agentdiag/internal/model"
)

func fixture() model.Report {
	return model.Report{ToolVersion: "0.1.0", GeneratedAt: time.Date(2026, 8, 19, 1, 2, 3, 0, time.UTC), Agents: []model.AgentReport{{Name: "Hermes", Detected: true, Home: "/tmp/.hermes", MCPServers: 2, Skills: []model.SkillInfo{{Name: "a"}}, Plugins: []model.PluginInfo{{Name: "p"}}, Findings: []model.Finding{{ID: "x", Severity: model.SeverityWarning, Message: "safe message", Security: true}}}}}
}

func TestRenderJSONAndMarkdown(t *testing.T) {
	for _, format := range []Format{FormatJSON, FormatMarkdown, FormatTerminal} {
		b, err := Render(fixture(), format)
		if err != nil {
			t.Fatalf("%s: %v", format, err)
		}
		s := string(b)
		if !strings.Contains(s, "Hermes") {
			t.Fatalf("%s missing agent: %s", format, s)
		}
		if strings.Contains(s, "super-secret") {
			t.Fatalf("%s leaked secret", format)
		}
	}
}
