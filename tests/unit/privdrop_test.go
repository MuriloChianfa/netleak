// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 MuriloChianfa

package unit

import (
	"os"
	"strings"
	"testing"
)

// buildChildEnv mirrors the logic in cmd/privdrop.go for unit testing.
// It clones the given environment, overrides identity/home vars, and
// optionally injects SSH_AUTH_SOCK via the provided discovery function.
func buildChildEnv(env []string, username, homeDir, uidStr string, discoverSock func(string) string) []string {
	overrides := map[string]string{
		"HOME":            homeDir,
		"USER":            username,
		"LOGNAME":         username,
		"XDG_RUNTIME_DIR": "/run/user/" + uidStr,
	}

	result := make([]string, 0, len(env)+len(overrides))
	seen := make(map[string]bool, len(overrides))
	hasAuthSock := false

	for _, kv := range env {
		key, _, _ := strings.Cut(kv, "=")
		if key == "SSH_AUTH_SOCK" {
			hasAuthSock = true
		}
		if v, ok := overrides[key]; ok {
			result = append(result, key+"="+v)
			seen[key] = true
			continue
		}
		result = append(result, kv)
	}

	for k, v := range overrides {
		if !seen[k] {
			result = append(result, k+"="+v)
		}
	}

	if !hasAuthSock && discoverSock != nil {
		if sock := discoverSock(uidStr); sock != "" {
			result = append(result, "SSH_AUTH_SOCK="+sock)
		}
	}

	return result
}

func envMap(env []string) map[string]string {
	m := make(map[string]string, len(env))
	for _, kv := range env {
		k, v, _ := strings.Cut(kv, "=")
		m[k] = v
	}
	return m
}

func TestBuildChildEnvOverridesHome(t *testing.T) {
	input := []string{
		"HOME=/root",
		"USER=root",
		"LOGNAME=root",
		"PATH=/usr/bin",
	}
	got := envMap(buildChildEnv(input, "alice", "/home/alice", "1000", nil))

	if got["HOME"] != "/home/alice" {
		t.Errorf("HOME = %q, want /home/alice", got["HOME"])
	}
	if got["USER"] != "alice" {
		t.Errorf("USER = %q, want alice", got["USER"])
	}
	if got["LOGNAME"] != "alice" {
		t.Errorf("LOGNAME = %q, want alice", got["LOGNAME"])
	}
	if got["PATH"] != "/usr/bin" {
		t.Errorf("PATH should be preserved, got %q", got["PATH"])
	}
}

func TestBuildChildEnvAddsXDGRuntimeDir(t *testing.T) {
	input := []string{"HOME=/root"}
	got := envMap(buildChildEnv(input, "bob", "/home/bob", "1001", nil))

	if got["XDG_RUNTIME_DIR"] != "/run/user/1001" {
		t.Errorf("XDG_RUNTIME_DIR = %q, want /run/user/1001", got["XDG_RUNTIME_DIR"])
	}
}

func TestBuildChildEnvPreservesExistingSSHAuthSock(t *testing.T) {
	input := []string{
		"HOME=/root",
		"SSH_AUTH_SOCK=/tmp/existing-agent.sock",
	}
	got := envMap(buildChildEnv(input, "alice", "/home/alice", "1000", nil))

	if got["SSH_AUTH_SOCK"] != "/tmp/existing-agent.sock" {
		t.Errorf("SSH_AUTH_SOCK = %q, want /tmp/existing-agent.sock", got["SSH_AUTH_SOCK"])
	}
}

func TestBuildChildEnvDiscoversSSHAuthSock(t *testing.T) {
	input := []string{"HOME=/root"}
	fakeDiscover := func(uid string) string {
		return "/run/user/" + uid + "/ssh-agent.socket"
	}
	got := envMap(buildChildEnv(input, "alice", "/home/alice", "1000", fakeDiscover))

	if got["SSH_AUTH_SOCK"] != "/run/user/1000/ssh-agent.socket" {
		t.Errorf("SSH_AUTH_SOCK = %q, want /run/user/1000/ssh-agent.socket", got["SSH_AUTH_SOCK"])
	}
}

func TestBuildChildEnvNoSSHAuthSockWhenDiscoveryFails(t *testing.T) {
	input := []string{"HOME=/root"}
	fakeDiscover := func(_ string) string { return "" }
	got := envMap(buildChildEnv(input, "alice", "/home/alice", "1000", fakeDiscover))

	if _, found := got["SSH_AUTH_SOCK"]; found {
		t.Errorf("SSH_AUTH_SOCK should not be set when discovery returns empty")
	}
}

func TestBuildChildEnvDoesNotDuplicateKeys(t *testing.T) {
	input := []string{
		"HOME=/root",
		"USER=root",
		"LOGNAME=root",
		"XDG_RUNTIME_DIR=/run/user/0",
		"FOO=bar",
	}
	result := buildChildEnv(input, "alice", "/home/alice", "1000", nil)

	counts := make(map[string]int)
	for _, kv := range result {
		k, _, _ := strings.Cut(kv, "=")
		counts[k]++
	}
	for k, c := range counts {
		if c > 1 {
			t.Errorf("key %q appears %d times, want 1", k, c)
		}
	}
}

func TestBuildChildEnvPreservesUnrelatedVars(t *testing.T) {
	input := []string{
		"HOME=/root",
		"LANG=en_US.UTF-8",
		"TERM=xterm-256color",
		"EDITOR=vim",
	}
	got := envMap(buildChildEnv(input, "alice", "/home/alice", "1000", nil))

	for _, k := range []string{"LANG", "TERM", "EDITOR"} {
		if _, ok := got[k]; !ok {
			t.Errorf("expected %s to be preserved", k)
		}
	}
}

// TestSUDOEnvVarsPresent verifies that the current process (when run
// under sudo) has the expected SUDO_* variables. This test is skipped
// when not running as root, since it can only pass under sudo.
func TestSUDOEnvVarsPresent(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip("not running as root; SUDO_* vars only present under sudo")
	}
	for _, k := range []string{"SUDO_UID", "SUDO_GID", "SUDO_USER"} {
		if os.Getenv(k) == "" {
			t.Errorf("%s is empty under sudo", k)
		}
	}
}
