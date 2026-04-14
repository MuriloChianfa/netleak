# SPDX-License-Identifier: MIT
# Copyright (c) 2024-2026 MuriloChianfa

ALPINE_VERSION  := 3.21
ALPINE_IMAGE    := alpine-cloud-$(ALPINE_VERSION).qcow2
ALPINE_URL      := https://dl-cdn.alpinelinux.org/alpine/v$(ALPINE_VERSION)/releases/cloud/nocloud_alpine-$(ALPINE_VERSION).0-x86_64-bios-cloudinit-r0.qcow2
SSH_PORT        := 2222
SSH_KEY         := .ssh/e2e_key
NETLEAK_BIN     := ../../netleak
NETLEAK_BPF     := ../../bpf/netleak.o

SCENARIOS := wireguard openvpn strongswan softether l2tp-ipsec multi-interface
BENCH_SCENARIOS := wireguard openvpn strongswan softether l2tp-ipsec

.PHONY: all download-image build-seed-iso ssh-keygen check-binary \
        run-all $(addprefix run-,$(SCENARIOS)) \
        bench-all $(addprefix bench-,$(BENCH_SCENARIOS)) aggregate-results charts clean

all: run-all

check-binary:
	@test -x $(NETLEAK_BIN) || { echo "Error: netleak binary not found. Run 'make' in repo root first."; exit 1; }
	@test -f $(NETLEAK_BPF) || { echo "Error: netleak.o not found. Run 'make' in repo root first."; exit 1; }

$(SSH_KEY):
	@mkdir -p .ssh
	ssh-keygen -t ed25519 -f $(SSH_KEY) -N "" -q

ssh-keygen: $(SSH_KEY)

$(ALPINE_IMAGE):
	@echo "Downloading Alpine cloud image $(ALPINE_VERSION)..."
	curl -fSL -o $(ALPINE_IMAGE) $(ALPINE_URL)

download-image: $(ALPINE_IMAGE)

seed.iso: $(SSH_KEY)
	@echo "Building cloud-init seed ISO..."
	@mkdir -p seed/
	@PUB=$$(cat $(SSH_KEY).pub) && \
	printf '#cloud-config\nssh_authorized_keys:\n  - %s\ndisable_root: false\nssh_pwauth: false\npackage_update: true\npackages:\n  - python3\nruncmd:\n  - passwd -u root\n  - mkdir -p /root/.ssh\n  - cp /home/alpine/.ssh/authorized_keys /root/.ssh/authorized_keys\n  - chown root:root /root/.ssh /root/.ssh/authorized_keys\n  - chmod 700 /root/.ssh\n  - chmod 600 /root/.ssh/authorized_keys\n  - sed -i "s/^#\\?PermitRootLogin.*/PermitRootLogin prohibit-password/" /etc/ssh/sshd_config\n  - rc-service sshd restart\n' "$$PUB" > seed/user-data
	@printf 'instance-id: netleak-e2e\nlocal-hostname: netleak-e2e\n' > seed/meta-data
	@genisoimage -output seed.iso -volid cidata -joliet -rock seed/meta-data seed/user-data 2>/dev/null \
	  || mkisofs -output seed.iso -volid cidata -joliet -rock seed/meta-data seed/user-data
	@rm -rf seed/

build-seed-iso: seed.iso

run-all: check-binary $(ALPINE_IMAGE) seed.iso $(SSH_KEY)
	@failed=0; \
	for scenario in $(SCENARIOS); do \
	  echo "========== Running e2e: $$scenario =========="; \
	  if ! bash scripts/boot-vm.sh $(ALPINE_IMAGE) seed.iso $(SSH_KEY) $(SSH_PORT) $$scenario; then \
	    echo "FAIL: $$scenario"; \
	    failed=$$((failed + 1)); \
	  else \
	    echo "PASS: $$scenario"; \
	  fi; \
	  echo; \
	done; \
	if [ $$failed -gt 0 ]; then \
	  echo "$$failed scenario(s) failed"; exit 1; \
	fi; \
	echo "All scenarios passed"

$(addprefix run-,$(SCENARIOS)): run-%: check-binary $(ALPINE_IMAGE) seed.iso $(SSH_KEY)
	bash scripts/boot-vm.sh $(ALPINE_IMAGE) seed.iso $(SSH_KEY) $(SSH_PORT) $*

bench-all: check-binary $(ALPINE_IMAGE) seed.iso $(SSH_KEY)
	@failed=0; \
	for scenario in $(BENCH_SCENARIOS); do \
	  echo "========== Running benchmark: $$scenario =========="; \
	  if ! bash scripts/boot-bench.sh $(ALPINE_IMAGE) seed.iso $(SSH_KEY) $(SSH_PORT) $$scenario; then \
	    echo "FAIL: bench-$$scenario"; \
	    failed=$$((failed + 1)); \
	  else \
	    echo "PASS: bench-$$scenario"; \
	  fi; \
	  echo; \
	done; \
	if [ $$failed -gt 0 ]; then \
	  echo "$$failed benchmark(s) failed"; exit 1; \
	fi; \
	echo "All benchmarks completed"

$(addprefix bench-,$(BENCH_SCENARIOS)): bench-%: check-binary $(ALPINE_IMAGE) seed.iso $(SSH_KEY)
	bash scripts/boot-bench.sh $(ALPINE_IMAGE) seed.iso $(SSH_KEY) $(SSH_PORT) $*

aggregate-results:
	bash scripts/aggregate-results.sh
	@$(MAKE) charts 2>/dev/null || echo "Chart generation skipped (install matplotlib: pip install -r requirements.txt)"

charts:
	python3 scripts/generate-charts.py

clean:
	rm -f $(ALPINE_IMAGE) seed.iso ../../netleak-static
	rm -rf seed/ .ssh/ logs/ results/
