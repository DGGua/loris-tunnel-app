package sshconfig

import (
	"os"
	"path/filepath"
	"testing"

	"loris-tunnel/internal/model"
)

func TestLoadImportCandidates(t *testing.T) {
	rootDir := t.TempDir()
	sshDir := filepath.Join(rootDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0o755); err != nil {
		t.Fatalf("mkdir ssh dir failed: %v", err)
	}

	includePath := filepath.Join(sshDir, "work.conf")
	if err := os.WriteFile(includePath, []byte(`
Host app-prod
  HostName 10.10.0.25
  User deploy
  Port 2200
  IdentityFile ~/.ssh/id_work
`), 0o644); err != nil {
		t.Fatalf("write include failed: %v", err)
	}

	configPath := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configPath, []byte(`
Host *
  User root
  ServerAliveInterval 15
  ConnectTimeout 8

Include work.conf

Host bastion
  HostName bastion.example.com
  User ops
  Port 2222
  IdentityAgent ~/.ssh/agent.sock
  StrictHostKeyChecking no

Host internal
  HostName internal.example.com
  ProxyJump bastion

Host *.corp
  User ignored
`), 0o644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	result, err := LoadImportCandidates(configPath)
	if err != nil {
		t.Fatalf("LoadImportCandidates failed: %v", err)
	}

	if len(result.Candidates) != 3 {
		t.Fatalf("Candidates len = %d, want 3", len(result.Candidates))
	}

	first := result.Candidates[0]
	if first.Alias != "app-prod" {
		t.Fatalf("first alias = %q, want app-prod", first.Alias)
	}
	if first.User != "deploy" {
		t.Fatalf("app-prod user = %q, want deploy", first.User)
	}
	if first.Port != 2200 {
		t.Fatalf("app-prod port = %d, want 2200", first.Port)
	}
	if first.KeyPath != "~/.ssh/id_work" {
		t.Fatalf("app-prod keyPath = %q, want ~/.ssh/id_work", first.KeyPath)
	}
	if first.KeepAliveIntervalMs != 15000 {
		t.Fatalf("app-prod keepAlive = %d, want 15000", first.KeepAliveIntervalMs)
	}
	if first.TimeoutMs != 8000 {
		t.Fatalf("app-prod timeout = %d, want 8000", first.TimeoutMs)
	}

	second := result.Candidates[1]
	if second.Alias != "bastion" {
		t.Fatalf("second alias = %q, want bastion", second.Alias)
	}
	if !second.BypassHostVerification {
		t.Fatalf("bastion bypassHostVerification = false, want true")
	}
	if second.AuthType != "ssh_agent" {
		t.Fatalf("bastion authType = %q, want ssh_agent", second.AuthType)
	}

	third := result.Candidates[2]
	if third.Alias != "internal" {
		t.Fatalf("third alias = %q, want internal", third.Alias)
	}
	if third.ProxyJump != "bastion" {
		t.Fatalf("internal proxyJump = %q, want bastion", third.ProxyJump)
	}
	if !contains(third.Warnings, "proxy_jump") {
		t.Fatalf("internal warnings = %v, want proxy_jump", third.Warnings)
	}
}

func TestLoadImportCandidatesMissingFile(t *testing.T) {
	_, err := LoadImportCandidates(filepath.Join(t.TempDir(), "missing-config"))
	if err == nil {
		t.Fatal("expected missing file error")
	}
}

func TestLoadImportCandidatesFallsBackToDefaultIdentity(t *testing.T) {
	rootDir := t.TempDir()
	t.Setenv("HOME", rootDir)
	t.Setenv("USERPROFILE", rootDir)

	sshDir := filepath.Join(rootDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0o755); err != nil {
		t.Fatalf("mkdir ssh dir failed: %v", err)
	}
	// Both default keys exist; id_ed25519 must win by preference order.
	// The public key file must never be picked.
	for _, name := range []string{"id_ed25519", "id_rsa", "id_rsa.pub"} {
		if err := os.WriteFile(filepath.Join(sshDir, name), []byte("fake"), 0o600); err != nil {
			t.Fatalf("write %s failed: %v", name, err)
		}
	}

	configPath := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configPath, []byte(`
Host nokey
  HostName nokey.example.com
  User ops

Host agenthost
  HostName agent.example.com
  IdentityAgent ~/.ssh/agent.sock

Host explicit
  HostName explicit.example.com
  IdentityFile ~/.ssh/custom
`), 0o644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	result, err := LoadImportCandidates(configPath)
	if err != nil {
		t.Fatalf("LoadImportCandidates failed: %v", err)
	}
	byAlias := make(map[string]model.SSHConfigImportCandidate)
	for _, candidate := range result.Candidates {
		byAlias[candidate.Alias] = candidate
	}

	nokey, ok := byAlias["nokey"]
	if !ok {
		t.Fatal("missing nokey candidate")
	}
	if nokey.KeyPath != "~/.ssh/id_ed25519" {
		t.Fatalf("nokey keyPath = %q, want ~/.ssh/id_ed25519", nokey.KeyPath)
	}
	if nokey.AuthType != "ssh_key" {
		t.Fatalf("nokey authType = %q, want ssh_key", nokey.AuthType)
	}
	if contains(nokey.Warnings, "agent_fallback") {
		t.Fatalf("nokey warnings = %v, want no agent_fallback", nokey.Warnings)
	}

	agenthost, ok := byAlias["agenthost"]
	if !ok {
		t.Fatal("missing agenthost candidate")
	}
	if agenthost.KeyPath != "" {
		t.Fatalf("agenthost keyPath = %q, want empty", agenthost.KeyPath)
	}
	if agenthost.AuthType != "ssh_agent" {
		t.Fatalf("agenthost authType = %q, want ssh_agent", agenthost.AuthType)
	}

	explicit, ok := byAlias["explicit"]
	if !ok {
		t.Fatal("missing explicit candidate")
	}
	if explicit.KeyPath != "~/.ssh/custom" {
		t.Fatalf("explicit keyPath = %q, want ~/.ssh/custom", explicit.KeyPath)
	}
}

func TestLoadImportCandidatesNoDefaultIdentity(t *testing.T) {
	rootDir := t.TempDir()
	t.Setenv("HOME", rootDir)
	t.Setenv("USERPROFILE", rootDir)

	sshDir := filepath.Join(rootDir, ".ssh")
	if err := os.MkdirAll(sshDir, 0o755); err != nil {
		t.Fatalf("mkdir ssh dir failed: %v", err)
	}

	configPath := filepath.Join(sshDir, "config")
	if err := os.WriteFile(configPath, []byte(`
Host nokey
  HostName nokey.example.com
  User ops
`), 0o644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	result, err := LoadImportCandidates(configPath)
	if err != nil {
		t.Fatalf("LoadImportCandidates failed: %v", err)
	}
	if len(result.Candidates) != 1 {
		t.Fatalf("Candidates len = %d, want 1", len(result.Candidates))
	}
	candidate := result.Candidates[0]
	if candidate.KeyPath != "" {
		t.Fatalf("keyPath = %q, want empty", candidate.KeyPath)
	}
	// ssh_key with an empty key path means "auto": try every private key
	// under ~/.ssh when the tunnel connects.
	if candidate.AuthType != "ssh_key" {
		t.Fatalf("authType = %q, want ssh_key", candidate.AuthType)
	}
	if contains(candidate.Warnings, "agent_fallback") {
		t.Fatalf("warnings = %v, want no agent_fallback", candidate.Warnings)
	}
}

func contains(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
