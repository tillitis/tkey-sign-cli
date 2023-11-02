tkey-sign(1)

# NAME

tkey-sign - Create or verify digital signatures with the Tillitis TKey

# SYNOPSIS

*tkey-sign* -h/--help

*tkey-sign* -G/--getkey -p/--public pubkey [-d/--port device] [-f/--force] [-s/--speed speed] [--uss] [--uss-file ussfile] [--verbose]

*tkey-sign* -S/--sign -m message -p/--public pubkey [-d/--port device] [-f/--force] [-s speed] [--uss] [--uss-file ussfile] [--verbose] [-x sigfile]

*tkey-sign* -V/--verify -m message -p/--public pubkey [--verbose] [-x sigfile]

*tkey-sign* --version

# DESCRIPTION

*tkey-sign* creates (-S) and verifies (-V) cryptographic signatures of
files. The signature is created by the *signer* device app running on
the Tillitis TKey. The signer is automatically loaded into the TKey
when signing or extracting the public key.

*-G, --getkey*

	Load device app (if not already loaded) and output its public key into
	a file. Pass the filename to store the public key in with *-p pubkey*.
	Combine with *--uss* or *--uss-file ussfile* to change the key pair.

*-S, --sign*

	Load device app (if not already loaded) and sign the specified message
	file (*-m message*). Store a signature in a file. Default signature
	file is message file + ".sig", specify another filename with *-x
	sigfile*. Also needs *-p pubkey* which is used for the comment in the
	sigfile and to verify that you really have the same public key as the
	loaded device app.

	Combine with *--uss* or *--uss-file ussfile* to change the key pair.

*-V, --verify*

	Verify the signature of message. Does not need a connected TKey.
	Needs *-m message* and *-p pubkey*. Default signature file is message.sig.
	Specify with *-x sigfile*.

*-d, --port device*

	Specify the device path for the TKey, typically something like
	*/dev/ttyACM0*. By default the device is auto-detected.

*-f, --force*

	Force writing signature and pubkey file, overwriting existing files.

*-h, --help*

	Output short help text.

*-m message*

	When signing, the file containing the message to sign.
	When verifying, the file containing the message to verify.

*-p, --public pubkey*

	Specify path to public key file.

*-s, --speed speed*

	Specify the speed in bits/second for the TKey device. Default is
	62500 b/s.

*--uss*

	Interactively enter a User Supplied Secret to be used when loading
	a device app to the TKey. This is used to create different key
	pairs. A warning is printed if the app is already loaded and this
	USS is not used.

*--uss-file secret*

	Use a User Supplied Secret in a file to be used when loading a
	device app to the TKey. This is used to create different key
	pairs. A warning is printed if the app is already loaded and this
	USS is not used. Use *-* to input from stdin.

*--verbose*

	Be very verbose about what is happening, including hexdumps of
	traffic to and from the TKey.

*--version*

	Output version information and quit.

*-x sigfile*

	The signature file to create or use with verification. The default
	when not using *-x* is *message.sig*.

# FORMATS

Signature files are expected to contain a one-line comment beginning
with "untrusted comment: ", then encoded in BASE64:

- 2 byte algorithm identifier, so far only "Ed" for Ed25519.
- 8 byte key number (unused by tkey-sign).
- 64 byte Ed25519 signature.

Pubkey files are expected to contain, a one-line comment beginning
with "untrusted comment: ", then encoded in BASE64:

- 2 byte algorithm identifier, so far only "Ed" for Ed25519.
- 8 byte key number (unused by tkey-sign).
- 32 byte Ed25519 public key.

# EXIT STATUS

The exit code is 0 on success and >0 if an error occurs.

It may fail because of one of the following reasons:

- Connection with or uploading of the app to the TKey failed.
- Some necessary files do not exist.
- File parse error.
- Public key of device app doesn't match public key file.
- The computed message file signature does not match the signature file.

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

All examples either load the device app automatically or works with an
already loaded device app.

Store the public key in a file. When loading the device app modify the
measured identity with a User Supplied Secret. The public key file `key.pub`
already exists so `tkey-sign` asks if it should replace it.

```
$ tkey-sign -G -p key.pub --uss
Enter phrase for the USS:
Repeat the phrase:
File key.pub exists. Overwrite [y/n]?
y
$
```

Sign a file. The program warns that the *signer* device app is already
loaded, probably from the above command.

```
./tkey-sign -S -m message -p key.pub
WARNING: App already loaded.
$
```

Verify a signature over a message file with the signature in the
default "message.sig" file:

```
$ tkey-sign -V -m message -p key.pub
Signature verified
$
```

# SEE ALSO

https://tillitis.se/
