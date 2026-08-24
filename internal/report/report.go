package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/TheJarv1s/agentdiag/internal/model"
)

type Format string

const (
	FormatTerminal Format = "terminal"
	FormatJSON     Format = "json"
	FormatMarkdown Format = "markdown"
)

func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "terminal", "text", "":
		return FormatTerminal, nil
	case "json":
		return FormatJSON, nil
	case "markdown", "md":
		return FormatMarkdown, nil
	default:
		return "", fmt.Errorf("unsupported format %q", s)
	}
}

func Render(r model.Report, format Format) ([]byte, error) {
	switch format {
	case FormatJSON:
		return json.MarshalIndent(r, "", "  ")
	case FormatMarkdown:
		return renderMarkdown(r), nil
	case FormatTerminal:
		return renderTerminal(r), nil
	default:
		return nil, fmt.Errorf("unsupported format %q", format)
	}
}

func renderTerminal(r model.Report) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "AgentDiag v%s\n", r.ToolVersion)
	if r.Project != "" {
		fmt.Fprintf(&b, "Project: %s\n", r.Project)
	}
	fmt.Fprintln(&b)
	for _, a := range r.Agents {
		status := "NOT DETECTED"
		if a.Detected {
			status = "DETECTED"
		}
		fmt.Fprintf(&b, "%s  [%s]\n", a.Name, status)
		if !a.Detected {
			fmt.Fprintln(&b)
			continue
		}
		fmt.Fprintf(&b, "  Home:    %s\n", a.Home)
		if a.ConfigPath != "" {
			fmt.Fprintf(&b, "  Config:  %s\n", a.ConfigPath)
		}
		fmt.Fprintf(&b, "  Skills:  %d\n  Plugins: %d\n  MCP:     %d\n", len(a.Skills), len(a.Plugins), a.MCPServers)
		if len(a.Findings) == 0 {
			fmt.Fprintln(&b, "  Findings: none")
		} else {
			fmt.Fprintf(&b, "  Findings: %d\n", len(a.Findings))
			for _, f := range a.Findings {
				fmt.Fprintf(&b, "    [%s] %s: %s", terminalFindingLabel(f), f.ID, f.Message)
				if f.Path != "" {
					fmt.Fprintf(&b, " (%s)", f.Path)
				}
				fmt.Fprintln(&b)
				if f.Remediation != "" {
					fmt.Fprintf(&b, "      Fix: %s\n", f.Remediation)
				}
			}
		}
		fmt.Fprintln(&b)
	}
	e, w, i := counts(r)
	fmt.Fprintf(&b, "Summary: %d error(s), %d warning(s), %d info\n", e, w, i)
	return b.Bytes()
}

func renderMarkdown(r model.Report) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "# AgentDiag report\n\n- **Tool:** v%s\n- **Generated:** %s\n", r.ToolVersion, r.GeneratedAt.Format("2006-01-02T15:04:05Z07:00"))
	if r.Project != "" {
		fmt.Fprintf(&b, "- **Project:** `%s`\n", markdownCode(r.Project))
	}
	e, w, i := counts(r)
	fmt.Fprintf(&b, "- **Summary:** %d error(s), %d warning(s), %d info\n\n", e, w, i)
	for _, a := range r.Agents {
		fmt.Fprintf(&b, "## %s\n\n", a.Name)
		if !a.Detected {
			fmt.Fprintln(&b, "Not detected.")
			fmt.Fprintln(&b)
			continue
		}
		fmt.Fprintf(&b, "- Home: `%s`\n- Skills: **%d**\n- Plugins: **%d**\n- MCP servers: **%d**\n\n", markdownCode(a.Home), len(a.Skills), len(a.Plugins), a.MCPServers)
		if len(a.Findings) == 0 {
			fmt.Fprintln(&b, "No findings.")
			fmt.Fprintln(&b)
			continue
		}
		fmt.Fprintln(&b, "### Findings")
		fmt.Fprintln(&b)
		findings := append([]model.Finding(nil), a.Findings...)
		sort.SliceStable(findings, func(i, j int) bool { return severityRank(findings[i].Severity) > severityRank(findings[j].Severity) })
		for _, f := range findings {
			fmt.Fprintf(&b, "- **%s — `%s`**: %s", markdownFindingLabel(f), f.ID, f.Message)
			if f.Path != "" {
				fmt.Fprintf(&b, " (`%s`)", markdownCode(f.Path))
			}
			if f.Remediation != "" {
				fmt.Fprintf(&b, "  \n  **Fix:** %s", f.Remediation)
			}
			fmt.Fprintln(&b)
		}
		fmt.Fprintln(&b)
	}
	fmt.Fprintln(&b, "---\n\nAgentDiag reports configuration structure and key names only; credential values are intentionally excluded.")
	return b.Bytes()
}

func terminalFindingLabel(f model.Finding) string {
	sev := strings.ToUpper(string(f.Severity))
	if f.Confidence == model.ConfidencePossible {
		return sev + "/POSSIBLE"
	}
	return sev
}

func markdownFindingLabel(f model.Finding) string {
	sev := strings.ToUpper(string(f.Severity))
	if f.Confidence == model.ConfidencePossible {
		return sev + " / POSSIBLE"
	}
	return sev
}

func markdownCode(s string) string { return strings.ReplaceAll(s, "`", "\\`") }

func counts(r model.Report) (errorsN, warningsN, infoN int) {
	for _, a := range r.Agents {
		for _, f := range a.Findings {
			switch f.Severity {
			case model.SeverityError:
				errorsN++
			case model.SeverityWarning:
				warningsN++
			default:
				infoN++
			}
		}
	}
	return
}

func severityRank(s model.Severity) int {
	switch s {
	case model.SeverityError:
		return 3
	case model.SeverityWarning:
		return 2
	default:
		return 1
	}
}
