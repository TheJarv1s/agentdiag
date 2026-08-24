package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/TheJarv1s/agentdiag/internal/engine"
	reportfmt "github.com/TheJarv1s/agentdiag/internal/report"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		args = []string{"scan"}
	}
	command := strings.ToLower(args[0])
	if command == "help" || command == "-h" || command == "--help" {
		printHelp(stdout)
		return 0
	}
	if command == "version" || command == "--version" || command == "-v" {
		fmt.Fprintf(stdout, "AgentDiag v%s\n", engine.Version)
		return 0
	}
	if command != "scan" && command != "doctor" && command != "security" && command != "export" {
		fmt.Fprintf(stderr, "unknown command %q\n\n", command)
		printHelp(stderr)
		return 2
	}

	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	fs.SetOutput(stderr)
	cwd, _ := os.Getwd()
	project := fs.String("project", cwd, "project directory to inspect")
	hermesHome := fs.String("hermes-home", "", "override Hermes home (default: HERMES_HOME or ~/.hermes)")
	ompHome := fs.String("omp-home", "", "override OMP data root (default: ~/.omp / XDG plugin data)")
	ompAgentDir := fs.String("omp-agent-dir", "", "override active OMP agent directory (default: PI_CODING_AGENT_DIR or <omp-home>/agent)")
	defaultFormat := "terminal"
	if command == "export" {
		defaultFormat = "markdown"
	}
	formatValue := fs.String("format", defaultFormat, "output format: terminal, json, markdown")
	output := fs.String("output", "", "output file (required by export)")
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}

	format, err := reportfmt.ParseFormat(*formatValue)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if command == "export" {
		if *output == "" {
			fmt.Fprintln(stderr, "export requires --output FILE")
			return 2
		}
		if format == reportfmt.FormatTerminal {
			fmt.Fprintln(stderr, "export supports --format markdown or json")
			return 2
		}
	}

	opts := engine.Options{Project: cleanProject(*project), HermesHome: *hermesHome, OMPHome: *ompHome, OMPAgentDir: *ompAgentDir}
	result := engine.Run(opts)
	mode := engine.ModeScan
	switch command {
	case "doctor":
		mode = engine.ModeDoctor
	case "security":
		mode = engine.ModeSecurity
	}
	result = engine.Filter(result, mode)
	rendered, err := reportfmt.Render(result, format)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	if command == "export" {
		target, err := filepath.Abs(*output)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		if err := os.WriteFile(target, append(rendered, '\n'), 0o600); err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		fmt.Fprintf(stdout, "Report written to %s\n", target)
	} else {
		if _, err := stdout.Write(rendered); err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		if len(rendered) == 0 || rendered[len(rendered)-1] != '\n' {
			fmt.Fprintln(stdout)
		}
	}
	if result.ErrorCount() > 0 {
		return 1
	}
	return 0
}

func cleanProject(p string) string {
	if strings.TrimSpace(p) == "" {
		return ""
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return filepath.Clean(p)
	}
	return abs
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, `AgentDiag — diagnostics for AI agent environments

Usage:
  agentdiag [scan] [flags]
  agentdiag doctor [flags]
  agentdiag security [flags]
  agentdiag export --format markdown|json --output FILE [flags]
  agentdiag version

Commands:
  scan       Detect Hermes/OMP and inventory skills, plugins and MCP (default)
  doctor     Run the full diagnostic report
  security   Show security-related findings only
  export     Write a sanitized Markdown or JSON support report
  version    Print AgentDiag version

Common flags:
  --project PATH        Project directory to inspect (default: current directory)
  --hermes-home PATH    Override Hermes home
  --omp-home PATH       Override OMP data root
  --omp-agent-dir PATH  Override active OMP agent directory
  --format FORMAT       terminal | json | markdown

AgentDiag v0.1 is read-only. It never executes plugins or agent binaries and never includes credential values in reports.`)
}
