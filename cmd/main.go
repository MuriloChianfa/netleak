// SPDX-License-Identifier: MIT
// Copyright (c) 2024-2026 MuriloChianfa

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/vishvananda/netlink"
)

const (
	fwmarkBase     = 0x4E4C0000 // "NL" prefix in upper 16 bits
	routeTableBase = 100
	routeTableMax  = 252
)

// allocFwmark returns a deterministic fwmark derived from the interface
// index. The upper 16 bits are the "NL" prefix; the lower 16 bits are
// the interface index (which is unique per-interface on a running system).
func allocFwmark(ifindex int) uint32 {
	idx := uint16(ifindex)
	return uint32(fwmarkBase) | uint32(idx)
}

// allocRouteTable returns a deterministic routing table ID derived from
// the interface index. The value stays within the valid user range [1, 252].
func allocRouteTable(ifindex int) int {
	span := routeTableMax - routeTableBase + 1
	idx := int(uint16(ifindex))
	return routeTableBase + (idx % span)
}

type ipMode int

const (
	ipModeAuto    ipMode = iota // dual-stack; blackhole v6 when interface lacks global v6
	ipModeV4Only                // route only IPv4, blackhole IPv6
	ipModeV6Only                // route only IPv6, blackhole IPv4
	ipModeFallback              // dual-stack; skip v6 rules when interface lacks global v6 (allows v6 leak)
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: netleak [options] <interface> <command> [args...]\n")
	fmt.Fprintf(os.Stderr, "       netleak --gateway <ip> <command> [args...]\n")
	fmt.Fprintf(os.Stderr, "       netleak --subnet <cidr> <command> [args...]\n\n")
	fmt.Fprintf(os.Stderr, "Routes all traffic from <command> (and children) through <interface>.\n")
	fmt.Fprintf(os.Stderr, "If the interface goes down, traffic is dropped (kill-switch).\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	fmt.Fprintf(os.Stderr, "  --gateway <ip>         Auto-detect interface by gateway IP\n")
	fmt.Fprintf(os.Stderr, "  --subnet <cidr>        Auto-detect interface by subnet (e.g. 10.10.0.0/24)\n")
	fmt.Fprintf(os.Stderr, "  --ingress-filter       Drop inbound packets not from the target interface\n")
	fmt.Fprintf(os.Stderr, "  --only-v4              Route only IPv4; block all IPv6 traffic\n")
	fmt.Fprintf(os.Stderr, "  --only-v6              Route only IPv6; block all IPv4 traffic\n")
	fmt.Fprintf(os.Stderr, "  --fallback-to-v4       If interface lacks IPv6, allow IPv6 via default route\n")
	fmt.Fprintf(os.Stderr, "  --keep-root            Run the command as root instead of dropping to the invoking user\n\n")
	fmt.Fprintf(os.Stderr, "Examples:\n")
	fmt.Fprintf(os.Stderr, "  sudo netleak ppp0 curl ifconfig.me\n")
	fmt.Fprintf(os.Stderr, "  sudo netleak --only-v4 ppp0 curl ifconfig.me\n")
	fmt.Fprintf(os.Stderr, "  sudo netleak --gateway 10.0.0.1 curl ifconfig.me\n")
	fmt.Fprintf(os.Stderr, "  sudo netleak --ingress-filter wg0 bash\n")
}

type cliOpts struct {
	iface         string
	cmdArgs       []string
	ingressFilter bool
	ipMode        ipMode
	keepRoot      bool
}

// parseArgs processes CLI arguments, resolving the target interface
// either by name or by --gateway/--subnet auto-detection, and collecting
// flags.
func parseArgs(args []string) (cliOpts, error) {
	var opts cliOpts

	i := 0
	for i < len(args) {
		switch args[i] {
		case "--ingress-filter":
			opts.ingressFilter = true
			i++

		case "--only-v4":
			opts.ipMode = ipModeV4Only
			i++

		case "--only-v6":
			opts.ipMode = ipModeV6Only
			i++

		case "--fallback-to-v4":
			opts.ipMode = ipModeFallback
			i++

		case "--keep-root":
			opts.keepRoot = true
			i++

		case "--gateway":
			if i+2 >= len(args) {
				return opts, fmt.Errorf("--gateway requires <ip> <command> [args...]")
			}
			iface, err := detectByGateway(args[i+1])
			if err != nil {
				return opts, fmt.Errorf("auto-detect by gateway: %w", err)
			}
			log.Printf("Auto-detected interface %s (gateway %s)", iface, args[i+1])
			opts.iface = iface
			i += 2

		case "--subnet":
			if i+2 >= len(args) {
				return opts, fmt.Errorf("--subnet requires <cidr> <command> [args...]")
			}
			iface, err := detectBySubnet(args[i+1])
			if err != nil {
				return opts, fmt.Errorf("auto-detect by subnet: %w", err)
			}
			log.Printf("Auto-detected interface %s (subnet %s)", iface, args[i+1])
			opts.iface = iface
			i += 2

		default:
			if args[i] == opts.iface {
				return opts, fmt.Errorf("interface %q already set by --gateway/--subnet; pass only the command: netleak --gateway <ip> <command>", args[i])
			}
			if opts.iface == "" {
				opts.iface = args[i]
				i++
			} else {
				opts.cmdArgs = args[i:]
				return opts, nil
			}
		}
	}

	if opts.iface == "" || len(opts.cmdArgs) == 0 {
		return opts, fmt.Errorf("not enough arguments")
	}
	return opts, nil
}

func main() {
	if len(os.Args) < 3 {
		usage()
		os.Exit(1)
	}

	opts, err := parseArgs(os.Args[1:])
	if err != nil {
		log.Fatalf("netleak: %v", err)
	}

	if len(opts.cmdArgs) == 0 {
		usage()
		os.Exit(1)
	}

	code, err := run(opts)
	if err != nil {
		log.Fatalf("netleak: %v", err)
	}
	os.Exit(code)
}

func run(opts cliOpts) (int, error) {
	targetIface := opts.iface
	cmdArgs := opts.cmdArgs

	// --- Verify target interface exists ---
	targetLink, err := netlink.LinkByName(targetIface)
	if err != nil {
		return 1, fmt.Errorf("interface %q: %w", targetIface, err)
	}
	ifindex := targetLink.Attrs().Index
	log.Printf("Target interface: %s (index %d)", targetIface, ifindex)

	mark := allocFwmark(ifindex)
	table := allocRouteTable(ifindex)

	// --- Register session ---
	pid := os.Getpid()
	if err := acquireSession(pid, targetIface, mark, table); err != nil {
		log.Printf("Warning: session tracking: %v", err)
	}
	defer releaseSession(pid)

	// --- Create cgroup v2 ---
	sessionID := strconv.Itoa(os.Getpid())
	cgPath, err := createCgroup(sessionID)
	if err != nil {
		return 1, err
	}
	defer cleanupCgroup(cgPath)
	log.Printf("Cgroup: %s", cgPath)

	// --- Load BPF programs ---
	objs, err := loadBPF()
	if err != nil {
		return 1, err
	}
	defer objs.Close()

	// --- Attach BPF programs to cgroup ---
	sockLink, egressLink, ingressLink, err := attachBPF(objs, cgPath, opts.ingressFilter)
	if err != nil {
		return 1, err
	}
	defer sockLink.Close()
	defer egressLink.Close()
	if ingressLink != nil {
		defer ingressLink.Close()
	}

	// --- Get cgroup ID ---
	cgID, err := getCgroupID(cgPath)
	if err != nil {
		return 1, fmt.Errorf("cgroup ID: %w", err)
	}
	log.Printf("Cgroup ID: %d", cgID)

	// --- Setup policy routing ---
	if err := setupRouting(targetLink, mark, table, opts.ipMode); err != nil {
		return 1, fmt.Errorf("routing: %w", err)
	}
	defer cleanupRouting(targetLink, mark, table)
	log.Printf("Policy routing: fwmark 0x%X -> table %d -> dev %s", mark, table, targetIface)

	// --- Populate BPF map: cgroup_id -> policy ---
	var polIfindex uint32
	if opts.ingressFilter {
		polIfindex = uint32(ifindex)
		log.Printf("Ingress filter: enabled (ifindex %d)", ifindex)
	}
	pol := policy{Fwmark: mark, Flags: 0, Ifindex: polIfindex}
	if err := objs.CgroupPolicyMap.Put(cgID, pol); err != nil {
		return 1, fmt.Errorf("populate map: %w", err)
	}
	defer func() {
		if err := objs.CgroupPolicyMap.Delete(cgID); err != nil {
			log.Printf("Warning: map cleanup: %v", err)
		}
	}()

	// --- Move self into cgroup ---
	if err := joinCgroup(cgPath, os.Getpid()); err != nil {
		return 1, fmt.Errorf("join cgroup: %w", err)
	}

	// --- Start interface monitor ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go monitorInterface(ctx, targetIface, objs.CgroupPolicyMap, cgID, mark, polIfindex)

	// --- Run target command ---
	log.Printf("Executing: %v", cmdArgs)
	return execAndWait(ctx, cancel, cmdArgs, opts.keepRoot)
}

// execAndWait forks the target command, forwards signals, waits for
// it to exit, and returns its exit code. Unless keepRoot is true, it
// drops privileges back to the invoking user when running under sudo.
func execAndWait(_ context.Context, cancel context.CancelFunc, cmdArgs []string, keepRoot bool) (int, error) {
	child := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	child.Stdin = os.Stdin
	child.Stdout = os.Stdout
	child.Stderr = os.Stderr

	if !keepRoot {
		if spa, env, err := dropPrivileges(); err == nil {
			child.SysProcAttr = spa
			child.Env = env
			log.Printf("Dropping privileges to uid=%d gid=%d",
				spa.Credential.Uid, spa.Credential.Gid)
		} else if !errors.Is(err, errNotSudo) {
			log.Printf("Warning: privilege drop failed: %v (running as root)", err)
		}
	}

	if err := child.Start(); err != nil {
		return 1, fmt.Errorf("exec %s: %w", cmdArgs[0], err)
	}

	// Forward signals to the child process
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range sigCh {
			_ = child.Process.Signal(sig)
		}
	}()

	// Wait for child to exit
	waitErr := child.Wait()
	signal.Stop(sigCh)
	close(sigCh)
	cancel()

	if waitErr != nil {
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			return exitErr.ExitCode(), nil
		}
		return 1, fmt.Errorf("wait: %w", waitErr)
	}
	return 0, nil
}
