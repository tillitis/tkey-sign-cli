# Check for OS
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
	shasum = shasum -a 512
	BUILD_CGO_ENABLED ?= 1
else
	shasum = sha512sum
	BUILD_CGO_ENABLED ?= 0
endif

.PHONY: all
all: check-signer-hash tkey-sign

.PHONY: windows
windows: tkey-sign.exe

DESTDIR=/
PREFIX=/usr/local
SYSTEMDDIR=/etc/systemd
UDEVDIR=/etc/udev
destbin=$(DESTDIR)/$(PREFIX)/bin
destman1=$(DESTDIR)/$(PREFIX)/share/man/man1
destunit=$(DESTDIR)/$(SYSTEMDDIR)/user
destrules=$(DESTDIR)/$(UDEVDIR)/rules.d
.PHONY: install
install:
	install -Dm755 tkey-sign $(destbin)/tkey-sign
	strip $(destbin)/tkey-sign
	install -Dm644 system/60-tkey.rules $(destrules)/60-tkey.rules
.PHONY: uninstall
uninstall:
	rm -f \
	$(destbin)/tkey-sign \
	$(destrules)/60-tkey.rules \

.PHONY: reload-rules
reload-rules:
	udevadm control --reload
	udevadm trigger

podman:
	podman run --rm --mount type=bind,source=$(CURDIR),target=/src --mount type=bind,source=$(CURDIR)/../tkey-libs,target=/tkey-libs -w /src -it ghcr.io/tillitis/tkey-builder:2 make -j

TKEY_SIGN_VERSION ?= $(shell git describe --dirty --always | sed -n "s/^v\(.*\)/\1/p")
# .PHONY to let go-build handle deps and rebuilds
.PHONY: tkey-sign
tkey-sign:
	CGO_ENABLED=$(BUILD_CGO_ENABLED) go build -ldflags "-X main.version=$(TKEY_SIGN_VERSION) -X main.signerAppNoTouch=$(TKEY_SIGNER_APP_NO_TOUCH)" -trimpath -o tkey-sign ./cmd/tkey-sign

.PHONY: tkey-sign.exe
tkey-sign.exe:
	$(MAKE) GOOS=windows GOARCH=amd64 tkey-sign

.PHONY: check-signer-hash
check-signer-hash:
	$(shasum) -c signer.bin.sha512

.PHONY: clean
clean:
	rm -f tkey-sign tkey-sign.exe

.PHONY: lint
lint:
	$(MAKE) -C gotools
	GOOS=linux   ./gotools/golangci-lint run
	GOOS=windows ./gotools/golangci-lint run
