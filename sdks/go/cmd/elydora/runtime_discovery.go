package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Elydora-Infra/Elydora-Go-SDK/v2/cmd/elydora/plugins"
)

type installedAgentRuntime struct {
	agentID   string
	agentName string
}

type runtimeIdentity struct {
	AgentID   string `json:"agent_id"`
	AgentName string `json:"agent_name"`
}

func readInstalledAgentRuntime(directoryName string) (*installedAgentRuntime, error) {
	agentDirectory, err := plugins.ResolveAgentRuntimeDirectory(directoryName)
	if err != nil {
		return nil, err
	}
	exists, err := plugins.RequirePhysicalDirectory(agentDirectory)
	if err != nil || !exists {
		return nil, err
	}

	configPath := filepath.Join(agentDirectory, "config.json")
	configExists, err := plugins.RequirePhysicalFile(configPath)
	if err != nil || !configExists {
		return nil, err
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read agent config %s: %w", configPath, err)
	}
	var identity runtimeIdentity
	if err := json.Unmarshal(data, &identity); err != nil {
		return nil, fmt.Errorf("parse agent config %s: %w", configPath, err)
	}
	if identity.AgentID == "" || identity.AgentName == "" {
		return nil, fmt.Errorf("agent config %s has an invalid runtime identity", configPath)
	}
	if identity.AgentID != directoryName {
		return nil, fmt.Errorf("agent config %s crosses its runtime directory", configPath)
	}
	if _, err := plugins.ResolveAgentRuntimeDirectory(identity.AgentID); err != nil {
		return nil, fmt.Errorf("validate agent config %s: %w", configPath, err)
	}
	return &installedAgentRuntime{
		agentID: identity.AgentID, agentName: identity.AgentName,
	}, nil
}

func discoverInstalledAgentRuntimes() ([]installedAgentRuntime, error) {
	runtimeRoot, err := plugins.AgentRuntimeRoot()
	if err != nil {
		return nil, err
	}
	exists, err := plugins.RequirePhysicalDirectory(runtimeRoot)
	if err != nil || !exists {
		return nil, err
	}
	entries, err := os.ReadDir(runtimeRoot)
	if err != nil {
		return nil, fmt.Errorf("read Elydora runtime root %s: %w", runtimeRoot, err)
	}

	runtimes := make([]installedAgentRuntime, 0, len(entries))
	for _, entry := range entries {
		if entry.Type()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf(
				"agent runtime path is not a physical directory: %s",
				filepath.Join(runtimeRoot, entry.Name()),
			)
		}
		if !entry.IsDir() {
			continue
		}
		runtime, err := readInstalledAgentRuntime(entry.Name())
		if err != nil {
			return nil, err
		}
		if runtime != nil {
			runtimes = append(runtimes, *runtime)
		}
	}
	return runtimes, nil
}

// findAgentIDByName returns one validated runtime identity for an agent name.
func findAgentIDByName(agentName string) (string, error) {
	runtimes, err := discoverInstalledAgentRuntimes()
	if err != nil {
		return "", err
	}
	foundAgentID := ""
	for _, runtime := range runtimes {
		if runtime.agentName != agentName {
			continue
		}
		if foundAgentID != "" {
			return "", fmt.Errorf(
				"multiple %q agent runtimes found; pass --agent-id explicitly",
				agentName,
			)
		}
		foundAgentID = runtime.agentID
	}
	return foundAgentID, nil
}

func resolveAgentRuntimeForUninstall(
	agentName, explicitAgentID string,
) (agentID, agentDirectory string, directoryExists bool, err error) {
	runtimeRoot, err := plugins.AgentRuntimeRoot()
	if err != nil {
		return "", "", false, err
	}
	runtimeRootExists, err := plugins.RequirePhysicalDirectory(runtimeRoot)
	if err != nil {
		return "", "", false, err
	}

	agentID = explicitAgentID
	if agentID == "" {
		agentID, err = findAgentIDByName(agentName)
		if err != nil || agentID == "" {
			return agentID, "", false, err
		}
	}
	agentDirectory, err = plugins.ResolveAgentRuntimeDirectory(agentID)
	if err != nil {
		return "", "", false, err
	}
	if runtimeRootExists {
		directoryExists, err = plugins.RequirePhysicalDirectory(agentDirectory)
		if err != nil {
			return "", "", false, err
		}
	}
	if directoryExists {
		runtime, err := readInstalledAgentRuntime(agentID)
		if err != nil {
			return "", "", false, err
		}
		if runtime != nil && runtime.agentName != agentName {
			return "", "", false, fmt.Errorf(
				"agent runtime %q belongs to %q, not %q",
				agentID, runtime.agentName, agentName,
			)
		}
	}
	return agentID, agentDirectory, directoryExists, nil
}
