package hermes

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/TheJarv1s/agentdiag/internal/fsx"
	"github.com/TheJarv1s/agentdiag/internal/model"
)

type Options struct {
	Home    string
	Project string
}

func Scan(opts Options) model.AgentReport {
	home := opts.Home
	if home == "" {
		home = os.Getenv("HERMES_HOME")
	}
	if home == "" {
		if h, err := os.UserHomeDir(); err == nil {
			home = filepath.Join(h, ".hermes")
		}
	}
	home, _ = filepath.Abs(home)
	r := model.AgentReport{Name: "Hermes", Home: home, ConfigPath: filepath.Join(home, "config.yaml")}
	r.Detected = fsx.Exists(r.ConfigPath) || fsx.IsDir(filepath.Join(home, "skills")) || fsx.IsDir(filepath.Join(home, "plugins")) || fsx.IsDir(filepath.Join(home, "hermes-agent"))
	if !r.Detected {
		if _, err := exec.LookPath("hermes"); err == nil {
			r.Detected = true
		}
	}
	if !r.Detected {
		return r
	}

	cfg := map[string]any{}
	if fsx.Exists(r.ConfigPath) {
		loaded, err := fsx.LoadYAML(r.ConfigPath)
		if err != nil {
			r.Findings = append(r.Findings, finding("hermes.config_invalid", model.SeverityError, "Hermes config.yaml could not be parsed.", r.ConfigPath, "Run `hermes config check` and repair the YAML.", false))
		} else {
			cfg = loaded
			for _, keyPath := range fsx.SecretKeyPaths(cfg) {
				r.Findings = append(r.Findings, finding("config.literal_secret", model.SeverityWarning, fmt.Sprintf("Credential-like value is stored literally at `%s` in config.yaml.", keyPath), r.ConfigPath, "Move the secret to ~/.hermes/.env and reference it via an environment variable.", true))
			}
			if m := fsx.MapAt(cfg, "mcp_servers"); m != nil {
				r.MCPServers = len(m)
			}
		}
	} else {
		r.Findings = append(r.Findings, finding("hermes.config_missing", model.SeverityInfo, "Hermes home was found but config.yaml is missing.", r.ConfigPath, "Run `hermes setup` or `hermes config check`.", false))
	}

	r.Skills = scanSkills(filepath.Join(home, "skills"), &r.Findings)
	r.Plugins = scanPlugins(filepath.Join(home, "plugins"), cfg, &r.Findings)
	checkSkillSafety(cfg, &r.Findings)
	checkEnvPermissions(filepath.Join(home, ".env"), &r.Findings)
	checkProjectPlugins(opts.Project, &r.Findings)
	sort.Slice(r.Findings, func(i, j int) bool {
		if r.Findings[i].Severity == r.Findings[j].Severity {
			return r.Findings[i].ID < r.Findings[j].ID
		}
		return severityRank(r.Findings[i].Severity) > severityRank(r.Findings[j].Severity)
	})
	return r
}

func scanSkills(root string, findings *[]model.Finding) []model.SkillInfo {
	files, err := fsx.FindSkillFiles(root, true)
	if err != nil {
		*findings = append(*findings, finding("hermes.skills_unreadable", model.SeverityError, "Hermes skills directory could not be scanned.", root, "Check filesystem permissions.", false))
		return nil
	}
	seen := map[string]string{}
	var out []model.SkillInfo
	for _, path := range files {
		meta, err := fsx.ParseSkill(path)
		if err != nil {
			*findings = append(*findings, finding("hermes.skill_invalid", model.SeverityWarning, "A Hermes SKILL.md has invalid frontmatter.", path, "Repair the YAML frontmatter.", false))
			continue
		}
		key := strings.ToLower(meta.Name)
		if prev, ok := seen[key]; ok {
			*findings = append(*findings, finding("hermes.skill_duplicate", model.SeverityWarning, fmt.Sprintf("Duplicate Hermes skill name `%s` shadows another skill.", meta.Name), path, fmt.Sprintf("Rename one skill; first seen at %s.", prev), false))
		} else {
			seen[key] = path
		}
		out = append(out, model.SkillInfo{Name: meta.Name, Description: meta.Description, Path: path, Source: "user"})
	}
	return out
}

func scanPlugins(root string, cfg map[string]any, findings *[]model.Finding) []model.PluginInfo {
	enabled := makeSet(fsx.StringSlice(fsx.ValueAt(cfg, "plugins", "enabled")))
	disabled := makeSet(fsx.StringSlice(fsx.ValueAt(cfg, "plugins", "disabled")))
	for name := range enabled {
		if disabled[name] {
			*findings = append(*findings, finding("hermes.plugin_state_conflict", model.SeverityWarning, fmt.Sprintf("Plugin `%s` appears in both plugins.enabled and plugins.disabled; disabled wins.", name), "", "Remove it from one list.", false))
		}
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		if !os.IsNotExist(err) {
			*findings = append(*findings, finding("hermes.plugins_unreadable", model.SeverityError, "Hermes plugins directory could not be scanned.", root, "Check filesystem permissions.", false))
		}
		return nil
	}
	var out []model.PluginInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(root, e.Name())
		manifest := filepath.Join(dir, "plugin.yaml")
		if !fsx.Exists(manifest) {
			continue
		}
		name := e.Name()
		if m, err := fsx.LoadYAML(manifest); err != nil {
			*findings = append(*findings, finding("hermes.plugin_manifest_invalid", model.SeverityWarning, "A Hermes plugin.yaml could not be parsed.", manifest, "Repair the plugin manifest.", false))
		} else if s, ok := m["name"].(string); ok && strings.TrimSpace(s) != "" {
			name = strings.TrimSpace(s)
		} else {
			*findings = append(*findings, finding("hermes.plugin_name_missing", model.SeverityWarning, "A Hermes plugin manifest has no name.", manifest, "Add a `name` field to plugin.yaml.", false))
		}
		p := model.PluginInfo{Name: name, Path: dir, Enabled: enabled[name], Disabled: disabled[name], Source: "user"}
		out = append(out, p)
		if !p.Enabled && !p.Disabled {
			*findings = append(*findings, finding("hermes.plugin_not_enabled", model.SeverityInfo, fmt.Sprintf("Plugin `%s` is installed but not enabled.", name), dir, fmt.Sprintf("Enable it with `hermes plugins enable %s` if trusted and needed.", name), false))
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func checkSkillSafety(cfg map[string]any, findings *[]model.Finding) {
	if v, ok := fsx.ValueAt(cfg, "skills", "guard_agent_created").(bool); ok && !v {
		*findings = append(*findings, finding("hermes.skill_guard_disabled", model.SeverityInfo, "Agent-created skill content guard is disabled.", "", "Consider enabling `skills.guard_agent_created` when accepting autonomous skill writes.", true))
	}
	if v, ok := fsx.ValueAt(cfg, "skills", "write_approval").(bool); ok && !v {
		*findings = append(*findings, finding("hermes.skill_write_approval_disabled", model.SeverityInfo, "Hermes can modify skills without per-write approval.", "", "Set `skills.write_approval: true` if you want every skill mutation staged for approval.", true))
	}
}

func checkEnvPermissions(path string, findings *[]model.Finding) {
	if !permissionCheckEnabled(runtime.GOOS) {
		return
	}
	if fsx.Exists(path) && fsx.WorldReadable(path) {
		*findings = append(*findings, finding("hermes.env_permissions", model.SeverityWarning, "Hermes .env is readable by group/other users.", path, "Restrict permissions to the current user (for example chmod 600 on Unix).", true))
	}
}

func checkProjectPlugins(project string, findings *[]model.Finding) {
	if project == "" || !strings.EqualFold(os.Getenv("HERMES_ENABLE_PROJECT_PLUGINS"), "true") {
		return
	}
	root := filepath.Join(project, ".hermes", "plugins")
	entries, err := os.ReadDir(root)
	if err == nil && len(entries) > 0 {
		*findings = append(*findings, finding("hermes.project_plugins_enabled", model.SeverityWarning, "Project-local Hermes plugins are enabled for a repository that contains plugin code.", root, "Only use HERMES_ENABLE_PROJECT_PLUGINS=true with trusted repositories.", true))
	}
}

func permissionCheckEnabled(goos string) bool { return goos != "windows" }

func makeSet(items []string) map[string]bool {
	out := map[string]bool{}
	for _, s := range items {
		out[s] = true
	}
	return out
}
func finding(id string, sev model.Severity, msg, path, remediation string, security bool) model.Finding {
	return model.Finding{ID: id, Severity: sev, Message: msg, Path: path, Remediation: remediation, Security: security}
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
