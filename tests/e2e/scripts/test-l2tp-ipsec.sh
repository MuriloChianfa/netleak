#!/usr/bin/env bash
# SPDX-License-Identifier: MIT
# Copyright (c) 2024-2026 MuriloChianfa
#
# L2TP/IPSec e2e test: verify netleak routes traffic through ppp0,
# kill-switch activates on interface down, and recovers on reconnect.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

SSH_KEY="$1"
SSH_PORT="$2"
IFACE="ppp0"
URL="http://10.10.0.1:8080/"

tap_begin

wait_for_http "${SSH_KEY}" "${SSH_PORT}" "${URL}"

# Test 1: Traffic routes through L2TP/IPSec PPP tunnel
assert_curl_ok "${SSH_KEY}" "${SSH_PORT}" "${IFACE}" "${URL}" \
    "traffic routes through ${IFACE}"

# Test 2: Kill-switch activates when ppp0 goes down
vm_ssh "${SSH_KEY}" "${SSH_PORT}" "ip link set ${IFACE} down"
sleep 2
assert_curl_fails "${SSH_KEY}" "${SSH_PORT}" "${IFACE}" "${URL}" \
    "kill-switch drops traffic when ${IFACE} is down"

# Test 3: Traffic recovers when L2TP tunnel is re-established
# PPP interfaces require session re-establishment rather than a simple link up
vm_ssh "${SSH_KEY}" "${SSH_PORT}" "bash -c '
    echo \"d vpn\" > /var/run/xl2tpd-client/l2tp-control 2>/dev/null || true
    sleep 1
    echo \"c vpn\" > /var/run/xl2tpd-client/l2tp-control
    for i in \$(seq 1 20); do
        if ip link show ppp0 2>/dev/null | grep -q UP; then
            exit 0
        fi
        sleep 1
    done
    exit 1
'"
sleep 4
assert_curl_ok "${SSH_KEY}" "${SSH_PORT}" "${IFACE}" "${URL}" \
    "traffic recovers when ${IFACE} is re-established"

tap_end
