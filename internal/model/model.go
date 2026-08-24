package model

import "time"

type Severity string

type Confidence string

const (
	ConfidenceConfirmed Confidence = "confirmed"
	ConfidencePossible  Confidence = "possible"
)

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

type Finding struct {
	ID          string     `json:"id"`
	Severity    Severity   `json:"severity"`
	Confidence  Confidence `json:"confidence,omitempty"`
	Message     string     `json:"message"`
	Path        string     `json:"path,omitempty"`
	Remediation string     `json:"remediation,omitempty"`
	Security    bool       `json:"security,omitempty"`
}

type SkillInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Path        string `json:"path"`
	Source      string `json:"source"`
	Managed     bool   `json:"managed,omitempty"`
}

type PluginInfo struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Enabled  bool   `json:"enabled,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
	Source   string `json:"source,omitempty"`
}

type AgentReport struct {
	Name       string       `json:"name"`
	Detected   bool         `json:"detected"`
	Home       string       `json:"home,omitempty"`
	ConfigPath string       `json:"config_path,omitempty"`
	Skills     []SkillInfo  `json:"skills,omitempty"`
	Plugins    []PluginInfo `json:"plugins,omitempty"`
	MCPServers int          `json:"mcp_servers"`
	Findings   []Finding    `json:"findings,omitempty"`
}

type Report struct {
	ToolVersion string        `json:"tool_version"`
	GeneratedAt time.Time     `json:"generated_at"`
	Project     string        `json:"project,omitempty"`
	Agents      []AgentReport `json:"agents"`
}

func (r Report) ErrorCount() int {
	n := 0
	for _, a := range r.Agents {
		for _, f := range a.Findings {
			if f.Severity == SeverityError {
				n++
			}
		}
	}
	return n
}
