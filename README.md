<img src=".github/banner.png" align="center"></img>

<p align="center">
  <a href="https://github.com/MuriloChianfa/netleak/releases/latest"><img src="https://img.shields.io/github/v/release/MuriloChianfa/netleak?label=release" alt="Release"></a>
  <a href="https://github.com/MuriloChianfa/netleak/actions/workflows/build.yml"><img src="https://github.com/MuriloChianfa/netleak/actions/workflows/build.yml/badge.svg" alt="Build"></a>
  <a href="https://github.com/MuriloChianfa/netleak/actions/workflows/e2e.yml"><img src="https://github.com/MuriloChianfa/netleak/actions/workflows/e2e.yml/badge.svg" alt="E2E VPN Tests"></a>
  <a href="https://github.com/MuriloChianfa/netleak/blob/main/LICENSE"><img src="https://img.shields.io/github/license/MuriloChianfa/netleak" alt="License"></a>
  <img src="https://img.shields.io/badge/platform-linux-blue" alt="Platform">
  <img src="https://img.shields.io/badge/go-1.24-00ADD8?logo=go&logoColor=white" alt="Go">
</p>

<p align="center">
  <img src="https://img.shields.io/badge/WireGuard-88171A?logo=wireguard&logoColor=white&style=for-the-badge" alt="WireGuard">
  <img src="https://img.shields.io/badge/OpenVPN-EA7E20?logo=openvpn&logoColor=white&style=for-the-badge" alt="OpenVPN">
  <img src="https://img.shields.io/badge/strongSwan-003399?style=for-the-badge" alt="strongSwan">
  <img src="https://img.shields.io/badge/SoftEther-0095D5?style=for-the-badge" alt="SoftEther">
</p>

<p>
  A kernel-level, proxychains-like tool for per-process traffic redirection on Linux. Built on top of <strong>cgroup v2</strong> and <strong>eBPF</strong>, netleak forces all network traffic from a process and its entire child tree through a specific network interface, completely transparently to the application. A kernel-enforced <strong>kill-switch</strong> drops every packet instead of falling back to the default route, guaranteeing zero traffic leakage under any failure conditions.
</p>

## Install from Package

<details name="install" open>
  <summary style="font-size: 16px;"><strong>Ubuntu/Debian</strong></summary>

  Download the `.deb` package from the [latest release](https://github.com/MuriloChianfa/netleak/releases/latest) and install it:

  ```bash
  curl -LO https://github.com/MuriloChianfa/netleak/releases/download/v1.1.1/netleak_1.1.1_amd64.deb
  sudo dpkg -i netleak_1.1.1_amd64.deb
  ```
</details>
<details name="install">
  <summary style="font-size: 16px;"><strong>Fedora/RHEL/Rocky/AlmaLinux</strong></summary>

  Download the `.rpm` package from the [latest release](https://github.com/MuriloChianfa/netleak/releases/latest) and install it:

  ```bash
  curl -LO https://github.com/MuriloChianfa/netleak/releases/download/v1.1.1/netleak-1.1.1-1.x86_64.rpm
  sudo rpm -i netleak-1.1.1-1.x86_64.rpm
  ```
</details>

## Or Build & Install from Source

<details name="build-source">
  <summary style="font-size: 16px;"><strong>Ubuntu/Debian</strong></summary>

  ```bash
  # Install build dependencies
  sudo apt update
  sudo apt install -y clang llvm libbpf-dev libelf-dev make golang

  # Clone and build
  git clone https://github.com/MuriloChianfa/netleak.git
  cd netleak
  make
  ```
</details>
<details name="build-source">
  <summary style="font-size: 16px;"><strong>Fedora/RHEL/Rocky Linux</strong></summary>

  ```bash
  # Install build dependencies
  sudo dnf install -y clang llvm libbpf-devel elfutils-libelf-devel make golang

  # Clone and build
  git clone https://github.com/MuriloChianfa/netleak.git
  cd netleak
  make
  ```
</details>

## Usage

```
sudo netleak <interface> <command> [args...]
```

### Examples

Route curl through a specific interface:

```bash
sudo netleak ppp0 curl ifconfig.me
```

Run a shell with all traffic going through a given interface:

```bash
sudo netleak wg0 bash
```

> [!IMPORTANT]
>
> Everything launched from that shell (and its children*), will have traffic routed through `wg0`.

## Security

For details on reporting vulnerabilities and our security practices, see the [Security Policy](https://github.com/MuriloChianfa/netleak/security/policy).

## License

The eBPF/C source code is licensed under GPL-3.0-only ([LICENSE-GPL](LICENSE-GPL)).

The Go source code is licensed under MIT ([LICENSE](LICENSE)).
