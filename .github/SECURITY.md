# Security Policy

## Supported Versions

The following versions are currently being supported with security updates:

| Version | Supported |
| ------- | --------- |
| 1.1.x   |     Yes   |
| 1.0.x   |     No    |
| 0.x.x   |     No    |

## Reporting a Vulnerability

The netleak team takes security vulnerabilities seriously. We appreciate your efforts to responsibly disclose your findings and will make every effort to acknowledge your contributions.

### How to Report a Security Vulnerability

**Please DO NOT report security vulnerabilities through public GitHub issues.**

Instead, please report them via email to:

**murilo.chianfa@outlook.com**

### What to Include in Your Report

To help us better understand the nature and scope of the security issue, please include as much of the following information as possible:

- **Type of vulnerability** (e.g., privilege escalation, BPF map manipulation, cgroup escape, etc.)
- **Full paths of source file(s)** related to the vulnerability
- **Location of the affected source code** (tag/branch/commit or direct URL)
- **Step-by-step instructions** to reproduce the issue
- **Proof-of-concept or exploit code** (if available)
- **Impact of the vulnerability**, including how an attacker might exploit it
- **Affected versions** of netleak
- **Kernel version and configuration** used during testing
- **Any special configuration** required to reproduce the issue

### Preferred Language

Please use **English** for all communications.

## Response Timeline

- **Initial Response**: We aim to acknowledge receipt of your vulnerability report within **48 hours**.
- **Status Updates**: We will send you regular updates about our progress, at least every **7 days**.
- **Disclosure Timeline**: We aim to patch critical vulnerabilities within **90 days** of the initial report.

## What to Expect

### After You Submit a Report

1. **Acknowledgment**: We will confirm receipt of your report within 48 hours
2. **Assessment**: We will assess the vulnerability and determine its severity
3. **Communication**: We will keep you informed of our progress
4. **Fix Development**: We will develop a patch for the vulnerability
5. **Testing**: We will test the fix thoroughly
6. **Disclosure**: We will coordinate with you on the disclosure timeline

### Severity Assessment

We use the following criteria to assess vulnerability severity:

- **Critical**: Privilege escalation, BPF program bypass, or kernel-level data corruption
- **High**: Kill-switch bypass, traffic leakage to unintended interfaces, cgroup escape
- **Medium**: Limited impact issues affecting specific kernel versions or configurations
- **Low**: Minor issues with minimal security impact

## Security Update Process

When we release a security fix:

1. **Private Patch**: We first create a private patch
2. **Notification**: We notify you and request validation of the fix
3. **Release**: We release the patch in a new version
4. **Advisory**: We publish a security advisory with details
5. **Credit**: We credit you in the advisory (unless you prefer to remain anonymous)

## Security Considerations for netleak

### eBPF and Kernel Security

- netleak loads eBPF programs into the kernel and **requires root privileges**
- BPF maps are pinned to `/sys/fs/bpf` and shared across sessions
- The kill-switch mechanism is enforced at the kernel level via cgroup-skb hooks

### Attack Surface

- **BPF Map Manipulation**: The cgroup policy map could be tampered with if bpffs permissions are misconfigured
- **Cgroup Escape**: Processes should not be able to leave the assigned cgroup
- **Routing Table Tampering**: Policy routing rules use a fixed fwmark and table ID
- **Kill-Switch Bypass**: The egress BPF program must always enforce the kill-switch when flagged

### Hardening Recommendations

1. Ensure `/sys/fs/bpf` has restrictive permissions
2. Run netleak only with the minimum required capabilities (CAP_SYS_ADMIN, CAP_NET_ADMIN, CAP_BPF)
3. Verify cgroup v2 is properly configured and mounted
4. Monitor kernel logs for unexpected BPF program detachments

## Security Advisories

Published security advisories can be found at:

- GitHub Security Advisories: https://github.com/MuriloChianfa/netleak/security/advisories
- Release Notes: https://github.com/MuriloChianfa/netleak/releases

## Bug Bounty Program

We do not currently have a bug bounty program, but we deeply appreciate security research and will publicly acknowledge your contributions (with your permission).

## Hall of Fame

We recognize security researchers who have helped improve netleak's security:

<!-- Security researchers will be listed here -->

_No security vulnerabilities have been reported yet._

## Questions?

If you have questions about this security policy, please email us at murilo.chianfa@outlook.com.

---

**Thank you for helping keep netleak and its users safe!**
