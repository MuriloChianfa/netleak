// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 MuriloChianfa

package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

var errNotSudo = errors.New("not running under sudo")

// dropPrivileges inspects SUDO_UID/SUDO_GID/SUDO_USER to build a
// syscall.SysProcAttr that drops the child process back to the
// invoking user, plus a corrected environment slice with the right
// HOME, USER, LOGNAME, and (best-effort) SSH_AUTH_SOCK.
//
// Returns errNotSudo when the SUDO_* variables are absent.
func dropPrivileges() (*syscall.SysProcAttr, []string, error) {
	uidStr := os.Getenv("SUDO_UID")
	gidStr := os.Getenv("SUDO_GID")
	if uidStr == "" || gidStr == "" {
		return nil, nil, errNotSudo
	}

	uid, err := strconv.ParseUint(uidStr, 10, 32)
	if err != nil {
		return nil, nil, fmt.Errorf("parse SUDO_UID %q: %w", uidStr, err)
	}
	gid, err := strconv.ParseUint(gidStr, 10, 32)
	if err != nil {
		return nil, nil, fmt.Errorf("parse SUDO_GID %q: %w", gidStr, err)
	}

	u, err := user.LookupId(uidStr)
	if err != nil {
		return nil, nil, fmt.Errorf("lookup uid %s: %w", uidStr, err)
	}

	groups, err := supplementaryGroups(u)
	if err != nil {
		return nil, nil, fmt.Errorf("supplementary groups: %w", err)
	}

	spa := &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid:    uint32(uid),
			Gid:    uint32(gid),
			Groups: groups,
		},
	}

	env := buildChildEnv(u, uidStr)
	return spa, env, nil
}

// supplementaryGroups returns the numeric group IDs for the user.
func supplementaryGroups(u *user.User) ([]uint32, error) {
	gids, err := u.GroupIds()
	if err != nil {
		return nil, err
	}
	out := make([]uint32, 0, len(gids))
	for _, g := range gids {
		id, err := strconv.ParseUint(g, 10, 32)
		if err != nil {
			continue
		}
		out = append(out, uint32(id))
	}
	return out, nil
}

// buildChildEnv clones the current environment and overrides the
// identity/home variables to match the original (non-root) user.
// It also attempts to recover SSH_AUTH_SOCK when sudo stripped it.
func buildChildEnv(u *user.User, uidStr string) []string {
	overrides := map[string]string{
		"HOME":            u.HomeDir,
		"USER":            u.Username,
		"LOGNAME":         u.Username,
		"XDG_RUNTIME_DIR": "/run/user/" + uidStr,
	}

	env := os.Environ()
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

	if !hasAuthSock {
		if sock := discoverSSHAuthSock(uidStr); sock != "" {
			result = append(result, "SSH_AUTH_SOCK="+sock)
		}
	}

	return result
}

// discoverSSHAuthSock tries well-known locations for the SSH agent
// socket belonging to the given UID. Returns "" if none found.
func discoverSSHAuthSock(uidStr string) string {
	uid, _ := strconv.ParseUint(uidStr, 10, 32)

	candidates := []string{
		"/run/user/" + uidStr + "/ssh-agent.socket",
		"/run/user/" + uidStr + "/keyring/ssh",
	}
	for _, p := range candidates {
		if isSocketOwnedBy(p, uint32(uid)) {
			return p
		}
	}

	// Scan /tmp/ssh-*/agent.* sockets owned by the original user.
	matches, _ := filepath.Glob("/tmp/ssh-*/agent.*")
	for _, p := range matches {
		if isSocketOwnedBy(p, uint32(uid)) {
			return p
		}
	}

	return ""
}

// isSocketOwnedBy checks that path is a Unix socket owned by uid.
func isSocketOwnedBy(path string, uid uint32) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}
	if fi.Mode().Type()&os.ModeSocket == 0 {
		return false
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return false
	}
	if st.Uid != uid {
		return false
	}
	// Quick dial check to make sure the socket is alive.
	c, err := net.Dial("unix", path)
	if err != nil {
		return false
	}
	c.Close()
	return true
}
