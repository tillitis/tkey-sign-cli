.PHONY: all
all: tkey-sign

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

# .PHONY to let go-build handle deps and rebuilds
.PHONY: tkey-sign
tkey-sign:
	go build -ldflags "-X main.signerAppNoTouch=$(TKEY_SIGNER_APP_NO_TOUCH)"

.PHONY: tkey-sign.exe
tkey-sign.exe:
	$(MAKE) GOOS=windows GOARCH=amd64 tkey-sign

.PHONY: clean
clean:
	rm -f tkey-sign

.PHONY: lint
lint:
	$(MAKE) -C gotools
	GOOS=linux   ./gotools/golangci-lint run
	GOOS=windows ./gotools/golangci-lint run
