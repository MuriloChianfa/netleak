GOARCH   ?= $(shell go env GOARCH)
GOOS     ?= linux

MULTIARCH := $(shell dpkg-architecture -qDEB_HOST_MULTIARCH 2>/dev/null || \
               if [ "$(GOARCH)" = "arm64" ]; then echo aarch64-linux-gnu; \
               else echo x86_64-linux-gnu; fi)
ARCH_INCLUDES := /usr/include/$(MULTIARCH)

BPF_CFLAGS := -O2 -g \
  -target bpf \
  -Wall -Wextra \
  -I/usr/include/bpf \
  -I$(ARCH_INCLUDES)

BPF_SRCS := bpf/netleak.c
BPF_OBJS := bpf/netleak.o
TARGET   := netleak
PREFIX   ?= /usr/local

.PHONY: all cmd clean install uninstall

all: cmd $(BPF_OBJS)

%.o: %.c
	clang $(BPF_CFLAGS) -c $< -o $@

cmd:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -C cmd -buildvcs=false -o ../$(TARGET)

install: all
	install -d $(DESTDIR)$(PREFIX)/bin
	install -m 755 $(TARGET) $(DESTDIR)$(PREFIX)/bin/$(TARGET)

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/$(TARGET)

clean:
	rm -f $(BPF_OBJS)
	rm -f $(TARGET)

e2e: all
	$(MAKE) -C tests/e2e run-all

e2e-wireguard: all
	$(MAKE) -C tests/e2e run-wireguard

e2e-openvpn: all
	$(MAKE) -C tests/e2e run-openvpn

e2e-strongswan: all
	$(MAKE) -C tests/e2e run-strongswan

e2e-softether: all
	$(MAKE) -C tests/e2e run-softether

e2e-l2tp-ipsec: all
	$(MAKE) -C tests/e2e run-l2tp-ipsec

bench: all
	$(MAKE) -C tests/e2e bench-all
	$(MAKE) -C tests/e2e aggregate-results

bench-wireguard: all
	$(MAKE) -C tests/e2e bench-wireguard

bench-openvpn: all
	$(MAKE) -C tests/e2e bench-openvpn

bench-strongswan: all
	$(MAKE) -C tests/e2e bench-strongswan

bench-softether: all
	$(MAKE) -C tests/e2e bench-softether

bench-l2tp-ipsec: all
	$(MAKE) -C tests/e2e bench-l2tp-ipsec

show:
	sudo bpftool prog show
	sudo bpftool map list

logs:
	echo 1 | sudo tee /sys/kernel/debug/tracing/tracing_on
	sudo cat /sys/kernel/debug/tracing/trace_pipe
