package engine

import (
	"time"

	"github.com/TheJarv1s/agentdiag/internal/adapters/hermes"
	"github.com/TheJarv1s/agentdiag/internal/adapters/omp"
	"github.com/TheJarv1s/agentdiag/internal/model"
)

const Version = "0.1.1"

type Options struct {
	Project     string
	HermesHome  string
	OMPHome     string
	OMPAgentDir string
}

type Mode string

const (
	ModeScan     Mode = "scan"
	ModeDoctor   Mode = "doctor"
	ModeSecurity Mode = "security"
)

func Run(opts Options) model.Report {
	return model.Report{
		ToolVersion: Version,
		GeneratedAt: time.Now().UTC(),
		Project:     opts.Project,
		Agents: []model.AgentReport{
			hermes.Scan(hermes.Options{Home: opts.HermesHome, Project: opts.Project}),
			omp.Scan(omp.Options{Home: opts.OMPHome, AgentDir: opts.OMPAgentDir, Project: opts.Project}),
		},
	}
}

func Filter(r model.Report, mode Mode) model.Report {
	if mode != ModeSecurity {
		return r
	}
	out := r
	out.Agents = make([]model.AgentReport, len(r.Agents))
	for i, a := range r.Agents {
		out.Agents[i] = a
		out.Agents[i].Findings = nil
		for _, f := range a.Findings {
			if f.Security {
				out.Agents[i].Findings = append(out.Agents[i].Findings, f)
			}
		}
	}
	return out
}
