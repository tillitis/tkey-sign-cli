module github.com/tillitis/tkey-sign

go 1.19

require (
	github.com/gen2brain/beeep v0.0.0-20230307103607-6e717729cb4f
	github.com/go-toast/toast v0.0.0-20190211030409-01e6764cf0a4
	github.com/spf13/pflag v1.0.5
	github.com/tillitis/tkeyclient v0.0.0-20230511144543-9ee035fb0288
	github.com/tillitis/tkeysign v0.0.0-20230511181826-bdde22885b71
	go.bug.st/serial v1.5.0
	golang.org/x/crypto v0.7.0
	golang.org/x/sys v0.8.0
	golang.org/x/term v0.7.0
)

require (
	github.com/creack/goselect v0.1.2 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d // indirect
	github.com/tadvi/systray v0.0.0-20190226123456-11a2b8fa57af // indirect
	github.com/tillitis/tkeyutil v0.0.0-00010101000000-000000000000 // indirect
)

replace github.com/tillitis/tkeyclient => ../tkeyclient

replace github.com/tillitis/tkeyutil => ../tkeyutil
