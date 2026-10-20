package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Elydora-Infra/Elydora-Go-SDK/v2/cmd/elydora/plugins"
)

const version = "2.1.1"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "install":
		cmdInstall(os.Args[2:])
	case "uninstall":
		cmdUninstall(os.Args[2:])
	case "status":
		cmdStatus(os.Args[2:])
	case "agents":
		cmdAgents()
	case "version", "--version", "-v":
		fmt.Printf("elydora %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Elydora CLI — agent hook installer for tamper-evident audit logging.

Usage:
  elydora <command> [flags]

Commands:
  install      Install Elydora audit hook for an agent
  uninstall    Remove Elydora audit hook for an agent
  status       Show installation status for an agent (or all agents)
  agents       List all supported agents
  version      Print version
  help         Show this help

Run "elydora <command> -h" for details on each command.
`)
}

func cmdInstall(args []string) {
	if err := rejectLegacySecretArguments(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fs := flag.NewFlagSet("install", flag.ExitOnError)
	agent := fs.String("agent", "", "Agent name (required). Use 'elydora agents' to list.")
	orgID := fs.String("org-id", "", "Organization ID (required)")
	agentID := fs.String("agent-id", "", "Agent ID (required)")
	privateKeyFile := fs.String("private-key-file", "", "Owner-only file containing the Ed25519 private key seed")
	kid := fs.String("kid", "", "Key ID (defaults to <agent-id>-key-1)")
	tokenFile := fs.String("token-file", "", "Owner-only file containing the optional API token")
	baseURL := fs.String("base-url", "https://api.elydora.com", "Elydora API base URL")

	fs.Parse(args)

	if *agent == "" || *orgID == "" || *agentID == "" {
		fmt.Fprintln(os.Stderr, "Error: --agent, --org-id, and --agent-id are required.")
		fmt.Fprintln(os.Stderr)
		fs.Usage()
		os.Exit(1)
	}

	if *kid == "" {
		*kid = *agentID + "-key-1"
	}

	plugin := plugins.NewPlugin(*agent)
	if plugin == nil {
		fmt.Fprintf(os.Stderr, "Error: unsupported agent %q. Run 'elydora agents' to see supported agents.\n", *agent)
		os.Exit(1)
	}
	agentDirectory, err := plugins.ResolveAgentRuntimeDirectory(*agentID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	secrets, err := resolveInstallSecrets(
		installSecretSources{
			privateKeyFile: *privateKeyFile,
			tokenFile:      *tokenFile,
		},
		newStandardSecretTerminal(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if err := validatePrivateKey(secrets.privateKey); err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid private key: %v\n", err)
		os.Exit(1)
	}

	guardScriptPath := filepath.Join(agentDirectory, "guard.js")
	config := plugins.InstallConfig{
		AgentName:       *agent,
		OrgID:           *orgID,
		AgentID:         *agentID,
		PrivateKey:      secrets.privateKey,
		KID:             *kid,
		Token:           secrets.token,
		BaseURL:         *baseURL,
		GuardScriptPath: guardScriptPath,
	}
	if preflighter, ok := plugin.(plugins.InstallPreflighter); ok {
		if err := preflighter.PreflightInstall(config); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	manager, managesGuard := plugin.(plugins.GuardRuntimeManager)
	if !managesGuard || !manager.ManagesGuardRuntime() {
		agentDirectory, err = plugins.PrepareAgentRuntimeDirectory(*agentID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		guardScript := plugins.GenerateGuardScript(*agent, *agentID)
		if err := plugins.WriteRuntimeFileAtomic(
			guardScriptPath,
			"Elydora guard runtime",
			[]byte(guardScript),
			0700,
		); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing guard script: %v\n", err)
			os.Exit(1)
		}
	}

	if err := plugin.Install(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  Guard script: %s\n", guardScriptPath)
}

func cmdUninstall(args []string) {
	fs := flag.NewFlagSet("uninstall", flag.ExitOnError)
	agent := fs.String("agent", "", "Agent name (required)")
	agentID := fs.String("agent-id", "", "Agent ID (if omitted, scans ~/.elydora/*/config.json for matching agent_name)")

	fs.Parse(args)

	if *agent == "" {
		fmt.Fprintln(os.Stderr, "Error: --agent is required.")
		fmt.Fprintln(os.Stderr)
		fs.Usage()
		os.Exit(1)
	}

	plugin := plugins.NewPlugin(*agent)
	if plugin == nil {
		fmt.Fprintf(os.Stderr, "Error: unsupported agent %q. Run 'elydora agents' to see supported agents.\n", *agent)
		os.Exit(1)
	}

	resolvedAgentID, agentDirectory, agentDirectoryExists, err := resolveAgentRuntimeForUninstall(*agent, *agentID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if resolvedAgentID == "" {
		fmt.Fprintf(os.Stderr, "Error: could not find agent ID for %q in ~/.elydora/*/config.json; pass --agent-id explicitly\n", *agent)
		os.Exit(1)
	}

	if err := plugin.Uninstall(resolvedAgentID); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Remove the per-agent directory
	if agentDirectoryExists {
		if err := os.RemoveAll(agentDirectory); err != nil {
			fmt.Fprintf(os.Stderr, "Error removing agent directory %s: %v\n", agentDirectory, err)
			os.Exit(1)
		}
		fmt.Printf("  Removed agent directory: %s\n", agentDirectory)
	}
}

func cmdStatus(args []string) {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	agent := fs.String("agent", "", "Agent name (optional — omit to show all)")

	fs.Parse(args)

	if *agent != "" {
		showAgentStatus(*agent)
		return
	}

	// Show status for all agents
	names := sortedAgentNames()
	for _, name := range names {
		showAgentStatus(name)
	}
}

func showAgentStatus(name string) {
	plugin := plugins.NewPlugin(name)
	if plugin == nil {
		fmt.Fprintf(os.Stderr, "Unknown agent: %s\n", name)
		return
	}

	status, err := plugin.Status()
	if err != nil {
		fmt.Fprintf(os.Stderr, "  %s: error getting status: %v\n", name, err)
		return
	}

	state := "not installed"
	if status.Installed {
		state = "installed"
	} else if status.HookConfigured && !status.HookScriptExists {
		state = "configured (hook script missing)"
	} else if !status.HookConfigured && status.HookScriptExists {
		state = "hook script exists (not configured)"
	}

	fmt.Printf("  %-14s  %-20s  %s\n", name, status.DisplayName, state)
}

func cmdAgents() {
	fmt.Println("Supported agents:")
	fmt.Println()
	fmt.Printf("  %-14s  %-20s  %s\n", "ID", "Name", "Config")
	fmt.Printf("  %-14s  %-20s  %s\n", strings.Repeat("-", 14), strings.Repeat("-", 20), strings.Repeat("-", 30))

	names := sortedAgentNames()
	for _, name := range names {
		entry := plugins.SupportedAgents[name]
		configPath := entry.ConfigDir + "/" + entry.ConfigFile // display only, not a filesystem path
		fmt.Printf("  %-14s  %-20s  %s\n", name, entry.Name, configPath)
	}
}

func sortedAgentNames() []string {
	names := make([]string, 0, len(plugins.SupportedAgents))
	for name := range plugins.SupportedAgents {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
