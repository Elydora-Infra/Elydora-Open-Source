package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Elydora-Infra/Elydora-Go-SDK/v2/cmd/elydora/plugins"
)

func setMainTestHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	return home
}

func TestGuardScriptPathRejectsUnsafeAgentID(t *testing.T) {
	setMainTestHome(t)
	_, err := plugins.ResolveAgentRuntimeDirectory("../escape")
	if err == nil || !strings.Contains(err.Error(), "invalid agent ID") {
		t.Fatalf("ResolveAgentRuntimeDirectory() error = %v", err)
	}
}

func TestFindAgentIDByNameReturnsValidatedDirectoryIdentity(t *testing.T) {
	home := setMainTestHome(t)
	runtimeDirectory := filepath.Join(home, ".elydora", "agent-1")
	if err := os.MkdirAll(runtimeDirectory, 0700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	config := []byte(`{"agent_name":"opencode","agent_id":"agent-1"}`)
	if err := os.WriteFile(filepath.Join(runtimeDirectory, "config.json"), config, 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	agentID, err := findAgentIDByName("opencode")
	if err != nil {
		t.Fatalf("findAgentIDByName() error = %v", err)
	}
	if agentID != "agent-1" {
		t.Fatalf("agent ID = %q, want agent-1", agentID)
	}
}

func TestFindAgentIDByNameRejectsConfigDirectoryMismatch(t *testing.T) {
	home := setMainTestHome(t)
	runtimeDirectory := filepath.Join(home, ".elydora", "stored-directory")
	if err := os.MkdirAll(runtimeDirectory, 0700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	config := []byte(`{"agent_name":"opencode","agent_id":"different-agent"}`)
	if err := os.WriteFile(filepath.Join(runtimeDirectory, "config.json"), config, 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := findAgentIDByName("opencode")
	if err == nil || !strings.Contains(err.Error(), "crosses its runtime directory") {
		t.Fatalf("findAgentIDByName() error = %v", err)
	}
}

func TestFindAgentIDByNameRejectsAmbiguousRuntimes(t *testing.T) {
	home := setMainTestHome(t)
	for _, agentID := range []string{"agent-1", "agent-2"} {
		runtimeDirectory := filepath.Join(home, ".elydora", agentID)
		if err := os.MkdirAll(runtimeDirectory, 0700); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		config := []byte(`{"agent_name":"opencode","agent_id":"` + agentID + `"}`)
		if err := os.WriteFile(filepath.Join(runtimeDirectory, "config.json"), config, 0600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
	}

	_, err := findAgentIDByName("opencode")
	if err == nil || !strings.Contains(err.Error(), "multiple") {
		t.Fatalf("findAgentIDByName() error = %v", err)
	}
}

func TestExplicitUninstallRejectsAgentOwnershipMismatch(t *testing.T) {
	home := setMainTestHome(t)
	runtimeDirectory := filepath.Join(home, ".elydora", "agent-1")
	if err := os.MkdirAll(runtimeDirectory, 0700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	config := []byte(`{"agent_name":"codex","agent_id":"agent-1"}`)
	if err := os.WriteFile(filepath.Join(runtimeDirectory, "config.json"), config, 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, _, _, err := resolveAgentRuntimeForUninstall("opencode", "agent-1")
	if err == nil || !strings.Contains(err.Error(), "belongs to") {
		t.Fatalf("resolveAgentRuntimeForUninstall() error = %v", err)
	}
}

func TestFindAgentIDByNameRejectsLinkedRuntime(t *testing.T) {
	home := setMainTestHome(t)
	root := filepath.Join(home, ".elydora")
	if err := os.MkdirAll(root, 0700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	target := t.TempDir()
	link := filepath.Join(root, "agent-1")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("directory symbolic links unavailable: %v", err)
	}

	_, err := findAgentIDByName("opencode")
	if err == nil || !strings.Contains(err.Error(), "physical directory") {
		t.Fatalf("findAgentIDByName() error = %v", err)
	}
}

func TestFindAgentIDByNameRejectsLinkedConfig(t *testing.T) {
	home := setMainTestHome(t)
	runtimeDirectory := filepath.Join(home, ".elydora", "agent-1")
	if err := os.MkdirAll(runtimeDirectory, 0700); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	target := filepath.Join(t.TempDir(), "config.json")
	config := []byte(`{"agent_name":"opencode","agent_id":"agent-1"}`)
	if err := os.WriteFile(target, config, 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.Symlink(target, filepath.Join(runtimeDirectory, "config.json")); err != nil {
		t.Skipf("file symbolic links unavailable: %v", err)
	}

	_, err := findAgentIDByName("opencode")
	if err == nil || !strings.Contains(err.Error(), "physical file") {
		t.Fatalf("findAgentIDByName() error = %v", err)
	}
}
