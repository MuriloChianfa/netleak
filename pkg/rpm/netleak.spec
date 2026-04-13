Name:           netleak
Version:        %{?version}%{!?version:1.1.1}
Release:        %{?release}%{!?release:1}%{?dist}
Summary:        cgroup-based eBPF per-process traffic redirection

License:        MIT AND GPL-3.0-only
URL:            https://github.com/MuriloChianfa/netleak
Source0:        %{name}-%{version}.tar.gz

Requires:       libbpf
Requires:       elfutils-libelf

# Copyright: 2024-2026 MuriloChianfa <murilo.chianfa@outlook.com>
# Go code is MIT licensed
# eBPF code is GPL-3.0-only licensed

%description
Route all traffic from a process (and its children) through an arbitrary
network interface, with kill-switch semantics that drop packets instead
of leaking when the interface goes down.

Features:
- Kernel-level proxychains-like behavior using cgroup v2 and eBPF
- Kill-switch: drops packets when the target interface is down
- No fallback routing under any failure condition
- Policy routing via firewall marks

%install
install -D -m 0755 %{_builddir}/netleak \
    %{buildroot}/usr/bin/netleak
install -D -m 0644 %{_builddir}/netleak.o \
    %{buildroot}/usr/lib/netleak/netleak.o
install -D -m 0644 %{_builddir}/LICENSE %{buildroot}%{_defaultdocdir}/%{name}/LICENSE
install -D -m 0644 %{_builddir}/LICENSE-GPL %{buildroot}%{_defaultdocdir}/%{name}/LICENSE-GPL

%files
%license %{_defaultdocdir}/%{name}/LICENSE
%license %{_defaultdocdir}/%{name}/LICENSE-GPL
%attr(0755, root, root) /usr/bin/netleak
%attr(0644, root, root) /usr/lib/netleak/netleak.o

%post
# Ensure BPF filesystem is mounted
if ! mountpoint -q /sys/fs/bpf 2>/dev/null; then
    echo "netleak: mounting BPF filesystem at /sys/fs/bpf"
    mount -t bpf bpf /sys/fs/bpf || true
fi

%preun
# Clean up pinned BPF maps
if [ -d /sys/fs/bpf ]; then
    rm -f /sys/fs/bpf/cgroup_policy_map 2>/dev/null || true
fi

# Clean up leftover netleak cgroups
if [ -d /sys/fs/cgroup/netleak ]; then
    rmdir /sys/fs/cgroup/netleak/*/ 2>/dev/null || true
    rmdir /sys/fs/cgroup/netleak    2>/dev/null || true
fi

%changelog
* Sun Apr 13 2026 MuriloChianfa <murilo.chianfa@outlook.com> - 1.1.1
- Release version 1.1.1
- Automatic privilege drop for child processes (SSH agent fix)
- New --keep-root flag to opt out of privilege dropping
- make install / make uninstall targets
- VPN throughput and latency benchmarks

* Thu Mar 26 2026 MuriloChianfa <murilo.chianfa@outlook.com> - 1.1.0
- Release version 1.1.0
- IPv6 policy routing support
- Multiple interface support
- Ingress filtering
- Interface auto-detection by gateway or IP range
- ARM64 and cross-compilation support

* Sun Feb 08 2026 MuriloChianfa <murilo.chianfa@outlook.com> - 1.0.0
- Release version 1.0.0
- Initial RPM package
