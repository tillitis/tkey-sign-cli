tkey-sign(1)

# NAME

tkey-sign - A hardware-backed digital signature command

# SYNOPSIS

*tkey-sign* sign [options...] FILE

*tkey-sign* verify [options...] FILE SIG-FILE PUBKEY-FILE

# DESCRIPTION

*tkey-sign* is a hardware-backed digital signature command using the
Tillitis TKey and the *signer* device app.

# OPTIONS

Common options for both commands:

*-h, --help*

	Output help text.

*--version*

	Output version information and quit.

# COMMANDS

## sign

*tkey-sign* sign [-p] [--port PATH] [--speed BPS] [--uss] [--uss-file
FILE] [--verbose]

*-p*, *--port PATH*

	Set serial port device PATH. If this is not passed, auto-detection
	will be attempted.


*--speed BPS*

	Set serial port speed to BPS b/s. Default is 62500 b/s.

*--uss*

	Ask for a phrase to be hashed as the User Supplied Secret. The
	USS is loaded onto the TKey along with the app itself. A
	different USS results in different different cryptographic key
	pair used for signing.

*--uss-file FILE*

	Read FILE and hash its contents as the USS. Use '-' (dash) to read
	from stdin. The full contents are hashed unmodified (e.g. newlines
	are not stripped).

*--verbose*

	Be more verbose, including reporting progress when writing to
	a file.

## verify

*tkey-sign* verify FILE SIG-FILE PUBKEY-FILE

Verifies the Ed25519 signature of FILE. Does not need a connected TKey
to verify.

SIG-FILE is expected to be 64 bytes Ed25519 signature in hex.
PUBKEY-FILE is expected to be 32 bytes Ed25519 public key in hex.

The exit code is 0 if the signature is valid, otherwise non-zero.
Newlines will be stripped from the input files.

# CONFIGURATION

You must have read and write access to the USB serial port TKey
provides. On platforms like macOS and Windows this happens
automatically when you approve the TKey device. Under Linux one way to
get access as your ordinary user is by installing a udev rule like
this:

```
# Mark Tillitis TKey as a security token. /usr/lib/udev/rules.d/70-uaccess.rules
# will add TAG "uaccess", which will result in file ACLs so that local user
# (see loginctl) can read/write to the serial port in /dev.
ATTRS{idVendor}=="1207", ATTRS{idProduct}=="8887",\
ENV{ID_SECURITY_TOKEN}="1"
```

Put this in */etc/udev/rules.d/60-tkey.rules* and run *udevadm control --reload*
which should make the TKey device (typically */dev/ttyACM0*)
availabe to anyone logged in on the console (see *loginctl*).

Another way to get access is by becoming a member of the group that
owns serial ports on some systems with default udev rules for USB CDC
ACM devices that come and go. On Ubuntu that group is *dialout*. You
can do it like this:

```
$ id -un
exampleuser
$ ls -l /dev/ttyACM0
crw-rw---- 1 root dialout 166, 0 Sep 16 08:20 /dev/ttyACM0
$ sudo usermod -a -G dialout exampleuser
```

For the change to take effect, you need to either log out and login
again or run the command *newgrp dialout* in the terminal that you are
working in.

# EXAMPLES

```
$ tkey-sign sign --uss Makefile
Auto-detected serial port /dev/ttyACM0
Connecting to TKey on serial port /dev/ttyACM0 ...
Enter phrase for the USS:
Repeat the phrase:
Signer app loaded.
Public Key from TKey: c06f2c8a0616b833640d450e56fe0cc04718569a5dac9c2b8c53ce55efd93f88
SHA512 hash: f0d89ef5c2594f5d9246128187e31bb99ee2f0631b5fad17293f3fff2a2b4f6238e52a282f28c6fd80a03ad2128183c6ded6e36faf3a1e4073a12cd6fb004681
The TKey will flash green when touch is required ...
Signature over message by TKey (on stdout):
7816315d8e746e935fe7399ccb323e343c9e17b4421501ce031bac577c914cc55413a6fc8526b3b0071de6a290fb2840d051e98286a81d23bd1068dce2179000
```

```
$ tkey-sign verify Makefile sig pubkey
Verifying signature ...
Public key: c06f2c8a0616b833640d450e56fe0cc04718569a5dac9c2b8c53ce55efd93f88
Signature: 7816315d8e746e935fe7399ccb323e343c9e17b4421501ce031bac577c914cc55413a6fc8526b3b0071de6a290fb2840d051e98286a81d23bd1068dce2179000
SHA512 hash: f0d89ef5c2594f5d9246128187e31bb99ee2f0631b5fad17293f3fff2a2b4f6238e52a282f28c6fd80a03ad2128183c6ded6e36faf3a1e4073a12cd6fb004681
Signature verified.
```

# SEE ALSO

https://tillitis.se/
