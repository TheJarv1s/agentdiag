package omp

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/TheJarv1s/agentdiag/internal/fsx"
	"github.com/TheJarv1s/agentdiag/internal/model"
)

type Options struct {
	Home     string
	AgentDir string
	Project  string
}

func Scan(opts Options) model.AgentReport {
	home := opts.Home
	if home == "" {
		if h, err := os.UserHomeDir(); err == nil {
			home = filepath.Join(h, ".omp")
		}
	}
	home, _ = filepath.Abs(home)
	agentDir := opts.AgentDir
	if agentDir == "" {
		agentDir = os.Getenv("PI_CODING_AGENT_DIR")
	}
	if agentDir == "" {
		agentDir = filepath.Join(home, "agent")
	}
	agentDir, _ = filepath.Abs(agentDir)

	configPath := filepath.Join(agentDir, "config.yml")
	if !fsx.Exists(configPath) && fsx.Exists(filepath.Join(agentDir, "config.yaml")) {
		configPath = filepath.Join(agentDir, "config.yaml")
	}
	pluginRoot := resolvePluginRoot(home, opts.Home != "")
	r := model.AgentReport{Name: "OMP", Home: home, ConfigPath: configPath}
	r.Detected = fsx.Exists(configPath) || fsx.IsDir(filepath.Join(agentDir, "skills")) || fsx.IsDir(filepath.Join(home, "plugins")) || fsx.IsDir(pluginRoot)
	if opts.Project != "" && fsx.IsDir(filepath.Join(opts.Project, ".omp")) {
		r.Detected = true
	}
	if !r.Detected {
		if _, err := exec.LookPath("omp"); err == nil {
			r.Detected = true
		}
	}
	if !r.Detected {
		return r
	}

	scanConfig(configPath, "OMP global config", &r.Findings)
	if opts.Project != "" {
		pconfig := filepath.Join(opts.Project, ".omp", "config.yml")
		if !fsx.Exists(pconfig) && fsx.Exists(filepath.Join(opts.Project, ".omp", "config.yaml")) {
			pconfig = filepath.Join(opts.Project, ".omp", "config.yaml")
		}
		if fsx.Exists(pconfig) {
			scanConfig(pconfig, "OMP project config", &r.Findings)
		}
	}

	var skills []model.SkillInfo
	authoredRoot := filepath.Join(agentDir, "skills")
	skills = append(skills, scanSkillRoot(authoredRoot, "user", false, &r.Findings)...)
	managedRoot := filepath.Join(agentDir, "managed-skills")
	skills = append(skills, scanSkillRoot(managedRoot, "managed", true, &r.Findings)...)
	if opts.Project != "" {
		skills = append(skills, scanSkillRoot(filepath.Join(opts.Project, ".omp", "skills"), "project", false, &r.Findings)...)
	}
	r.Skills = skills
	checkDuplicateSkills(r.Skills, &r.Findings)

	for _, path := range ompMCPPaths(agentDir, opts.Project) {
		if !fsx.Exists(path) {
			continue
		}
		n, err := fsx.CountMCPServersJSON(path)
		if err != nil {
			r.Findings = append(r.Findings, finding("omp.mcp_invalid", model.SeverityError, "An OMP MCP configuration file is invalid JSON.", path, "Repair the JSON; OMP cannot load servers from this file.", false))
			continue
		}
		r.MCPServers += n
	}

	r.Plugins = append(r.Plugins, scanPluginLock(filepath.Join(pluginRoot, "omp-plugins.lock.json"), "user", &r.Findings)...)
	if opts.Project != "" {
		r.Plugins = append(r.Plugins, scanPluginLock(filepath.Join(opts.Project, ".omp", "plugins", "omp-plugins.lock.json"), "project", &r.Findings)...)
	}
	scanMarketplaceRegistry(filepath.Join(pluginRoot, "installed_plugins.json"), &r.Findings)
	if opts.Project != "" {
		scanMarketplaceRegistry(filepath.Join(opts.Project, ".omp", "plugins", "installed_plugins.json"), &r.Findings)
	}

	sort.Slice(r.Skills, func(i, j int) bool {
		if r.Skills[i].Source == r.Skills[j].Source {
			return r.Skills[i].Name < r.Skills[j].Name
		}
		return r.Skills[i].Source < r.Skills[j].Source
	})
	sort.Slice(r.Plugins, func(i, j int) bool {
		if r.Plugins[i].Source == r.Plugins[j].Source {
			return r.Plugins[i].Name < r.Plugins[j].Name
		}
		return r.Plugins[i].Source < r.Plugins[j].Source
	})
	sort.Slice(r.Findings, func(i, j int) bool {
		if r.Findings[i].Severity == r.Findings[j].Severity {
			return r.Findings[i].ID < r.Findings[j].ID
		}
		return severityRank(r.Findings[i].Severity) > severityRank(r.Findings[j].Severity)
	})
	return r
}

func scanConfig(path, label string, findings *[]model.Finding) {
	if !fsx.Exists(path) {
		return
	}
	cfg, err := fsx.LoadYAML(path)
	if err != nil {
		f := finding("omp.config_unparsed", model.SeverityWarning, "AgentDiag could not fully parse "+label+" with its lightweight compatibility parser; this does not prove the config is invalid.", path, "Validate the configuration with OMP; AgentDiag will skip checks that depend on parsed values.", false)
		f.Confidence = model.ConfidencePossible
		*findings = append(*findings, f)
		return
	}
	for _, keyPath := range fsx.SecretKeyPaths(cfg) {
		*findings = append(*findings, finding("config.literal_secret", model.SeverityWarning, fmt.Sprintf("Credential-like value is stored literally at `%s` in OMP config.", keyPath), path, "Move the credential to an environment variable and reference it from configuration.", true))
	}
}

func scanSkillRoot(root, source string, managed bool, findings *[]model.Finding) []model.SkillInfo {
	direct, err := fsx.FindSkillFiles(root, false)
	if err != nil {
		*findings = append(*findings, finding("omp.skills_unreadable", model.SeverityError, "An OMP skills directory could not be scanned.", root, "Check filesystem permissions.", false))
		return nil
	}
	all, _ := fsx.FindSkillFiles(root, true)
	directSet := map[string]bool{}
	for _, p := range direct {
		directSet[filepath.Clean(p)] = true
	}
	for _, p := range all {
		if !directSet[filepath.Clean(p)] {
			*findings = append(*findings, finding("omp.skill_nested_ignored", model.SeverityWarning, "Nested OMP skill layout is outside native provider discovery depth and will be ignored.", p, "Move the skill to `<skills-root>/<skill-name>/SKILL.md` or configure the nested parent explicitly.", false))
		}
	}
	var out []model.SkillInfo
	for _, path := range direct {
		meta, err := fsx.ParseSkill(path)
		if err != nil {
			*findings = append(*findings, finding("omp.skill_unreadable", model.SeverityWarning, "An OMP SKILL.md could not be read.", path, "Check filesystem permissions and file encoding.", false))
			continue
		}
		if strings.TrimSpace(meta.Description) == "" {
			*findings = append(*findings, finding("omp.skill_missing_description", model.SeverityWarning, fmt.Sprintf("Native OMP skill `%s` has no description and may not be discovered by providers requiring one.", meta.Name), path, "Add a concise `description` to SKILL.md frontmatter.", false))
		}
		out = append(out, model.SkillInfo{Name: meta.Name, Description: meta.Description, Path: path, Source: source, Managed: managed})
	}
	return out
}

func checkDuplicateSkills(skills []model.SkillInfo, findings *[]model.Finding) {
	seen := map[string]model.SkillInfo{}
	for _, s := range skills {
		key := strings.ToLower(s.Name)
		if prev, ok := seen[key]; ok {
			*findings = append(*findings, finding("omp.skill_duplicate", model.SeverityWarning, fmt.Sprintf("OMP skill `%s` exists in multiple scanned sources (%s and %s).", s.Name, prev.Source, s.Source), s.Path, "Review provider precedence and keep only the intended copy when possible.", false))
		} else {
			seen[key] = s
		}
	}
}

func ompMCPPaths(agentDir, project string) []string {
	paths := []string{filepath.Join(agentDir, "mcp.json"), filepath.Join(agentDir, ".mcp.json")}
	if project != "" {
		paths = append([]string{filepath.Join(project, ".omp", "mcp.json"), filepath.Join(project, ".omp", ".mcp.json")}, paths...)
	}
	return paths
}

func resolvePluginRoot(home string, explicitHome bool) string {
	if !explicitHome {
		if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
			candidate := filepath.Join(xdg, "omp", "plugins")
			if fsx.IsDir(candidate) {
				return candidate
			}
		}
	}
	return filepath.Join(home, "plugins")
}

func scanPluginLock(path, source string, findings *[]model.Finding) []model.PluginInfo {
	if !fsx.Exists(path) {
		return nil
	}
	obj, err := fsx.LoadJSON(path)
	if err != nil {
		*findings = append(*findings, finding("omp.plugin_lock_invalid", model.SeverityWarning, "OMP plugin runtime lock is invalid JSON.", path, "Repair or regenerate plugin state with OMP plugin commands.", false))
		return nil
	}
	plugins, _ := obj["plugins"].(map[string]any)
	var out []model.PluginInfo
	for name, raw := range plugins {
		enabled := true
		if m, ok := raw.(map[string]any); ok {
			if v, ok := m["enabled"].(bool); ok {
				enabled = v
			}
		}
		out = append(out, model.PluginInfo{Name: name, Path: path, Enabled: enabled, Disabled: !enabled, Source: source})
	}
	return out
}

func scanMarketplaceRegistry(path string, findings *[]model.Finding) {
	if !fsx.Exists(path) {
		return
	}
	if _, err := fsx.LoadJSON(path); err != nil {
		*findings = append(*findings, finding("omp.marketplace_registry_invalid", model.SeverityWarning, "OMP marketplace installed_plugins.json is invalid JSON.", path, "Repair or reinstall marketplace plugin state.", false))
	}
}

func finding(id string, sev model.Severity, msg, path, remediation string, security bool) model.Finding {
	return model.Finding{ID: id, Severity: sev, Confidence: model.ConfidenceConfirmed, Message: msg, Path: path, Remediation: remediation, Security: security}
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
