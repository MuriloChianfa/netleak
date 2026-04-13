# netleak Roadmap

## v1.0.0 - Testing and CI

- [x] Unit tests for Go code (cgroup, routing, BPF loading)
- [x] Fuzz tests for argument parsing and edge cases
- [x] CI pipelines (multi-distro build matrix)
- [x] CodeQL security scanning (Go + C)
- [x] `.deb` and `.rpm` packaging
- [x] Automated release workflow (tag-triggered)
- [x] Dockerfile with multi-stage build

## v1.0.1 - Integration Testing (QEMU/KVM)

- [x] QEMU/KVM test harness
  - [x] VM boot scripts (QEMU launch with KVM auto-detection, TCG fallback)
  - [x] Alpine cloud image download and cloud-init seed ISO generation
  - [x] SSH-based VM access via user-mode networking port forward
  - [x] Ansible provisioning for declarative, idempotent VM configuration
  - [x] Shared test helpers and TAP-style output reporting
- [x] WireGuard integration test
  - [x] Ansible role: namespace + veth underlay, keypair generation, `wg0` tunnel
  - [x] Traffic routing verification (`netleak wg0 curl ...`)
  - [x] Kill-switch test (interface down -> traffic dropped)
  - [x] Recovery test (interface up -> traffic resumes)
- [x] OpenVPN integration test
  - [x] Ansible role: ephemeral PKI (CA + certs), TLS tunnel with `--dev tun`
  - [x] Traffic routing verification (`netleak tun0 curl ...`)
  - [x] Kill-switch and recovery validation
- [x] strongSwan/IPsec integration test
  - [x] Ansible role: IKEv2 with PSK, route-based VTI (`vti0` or `xfrm`) interface
  - [x] Traffic routing verification (`netleak vti0 curl ...`)
  - [x] Kill-switch and recovery validation
- [x] SoftEther VPN integration test
  - [x] Ansible role: SoftEther server with virtual hub, `tap0` bridge interface
  - [x] Traffic routing verification (`netleak tap0 curl ...`)
  - [x] Kill-switch and recovery validation (layer-2 tap device)
- [x] GitHub Actions CI workflow for e2e tests (TCG mode, parallel job matrix)
- [x] Test artifact collection and log upload

## v1.1.0 - IPv6 and Multi-Interface

- [x] IPv6 policy routing support
- [x] Multiple interface support (route different cgroups through different interfaces)
- [x] Ingress filtering (drop inbound traffic not from the target interface)
- [x] Interface auto-detection by gateway or IP range
- [x] Stable CLI and BPF map ABI
- [x] ARM64 and cross-compilation support

## v1.1.1 - VPN Throughput Benchmarks

- [x] Benchmark harness (QEMU/KVM, reuses e2e VM infrastructure, for both x86 and arm)
  - [x] `iperf3` server/client inside VPN tunnel namespaces
  - [x] Baseline measurement: `iperf3` direct through VPN interface (no netleak)
  - [x] Netleak measurement: `netleak <iface> iperf3 -c ...`
  - [x] Automated comparison and delta-percentage calculation
- [x] WireGuard (`wg0`) throughput benchmark
- [x] OpenVPN (`tun0`) throughput benchmark
- [x] strongSwan/IPsec (`vti0`) throughput benchmark
- [x] SoftEther (`tap0`) throughput benchmark
- [x] Latency benchmarks (RTT via `ping` with and without netleak)
- [x] Multi-stream benchmarks (parallel `iperf3` sessions)
- [x] Results output in machine-readable format (JSON/CSV)
- [x] CI job: run benchmarks and upload results as artifacts
- [x] Historical tracking: compare against previous runs to detect regressions
- [x] Automatic privilege drop for child processes when running under sudo
- [x] `--keep-root` flag to opt out of privilege dropping
- [x] `make install` / `make uninstall` targets

## v1.2.0 - Daemon Mode

- [ ] Daemon mode (`netleakd`) with D-Bus or Unix socket API
- [ ] NetworkManager plugin integration
- [ ] Systemd service unit
- [ ] Per-user and per-application policies
- [ ] Session persistence across reboots
- [ ] Configuration file support (TOML or YAML)
- [ ] Man pages (`netleak(8)`)
- [ ] Full documentation site
