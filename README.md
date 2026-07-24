
[![ci](https://github.com/tillitis/tkey-sign/actions/workflows/ci.yaml/badge.svg?branch=main&event=push)](https://github.com/tillitis/tkey-sign/actions/workflows/ci.yaml)

# tkey-sign

`tkey-sign` creates and verifies cryptographic signatures of files.
The signature is created by the [signer device
app](https://github.com/tillitis/tkey-device-signer) running on the
[Tillitis](https://tillitis.se/) TKey. The signer is automatically
loaded into the TKey by `tkey-sign` when signing or extracting the
public key. The measured private key never leaves the TKey.

See [Release notes](RELEASE.md).

## Usage

Run with `-h`. See examples below. For details, see manual page in
`doc/tkey-sign.1` and source in
[doc/tkey-sign.scd](doc/tkey-sign.scd).

The key and signature files are compatible with OpenBSD's `signify(1)`
but the way the signing works is not. Instead of signing the entire
file `tkey-sign` signs a digest of the message. The hash algorithm can
be SHA-512 (default) and BLAKE2s.

In order to be at least slightly compatible with `signify`,
`tkey-sign` when using the SHA-512 digest, does it in a way that is
compatible with:

```
sha512sum message-file >digestfile
signify -V -m digestfile -x message-file.sig -p key.pub
```

See the helper script `signify-verify`.

NB! This is slightly worrisome since it potentially includes the path
(not necessarily just basename) of the `message-file` into the message
that is signed depending on how you call `sha512sum`.

There is no similar support for BLAKE2s digests.

## Examples

All examples either load the device app automatically or works with an
already loaded device app.

Store the public key in a file.

```
$ tkey-sign -G -p key.pub
```

Sign a file using the signer's basic secret or the identity of an
already loaded signer while also checking that you have the right
public key in a file:

```
$ tkey-sign -S -m message.txt -p key.pub
```

Verify a signature over a message file with the signature in the
default "message.txt.sig" file:

```
$ tkey-sign -V -p key.pub -m message.txt
```

or

```
$ signify-verify message.txt key.pub
Signature Verified
$
```

## Build & install

The easiest way is to:

```
$ go install github.com/tillitis/tkey-sign-cli/cmd/tkey-sign@latest
```

After this the `tkey-sign` command should be available in your
`$GOBIN` directory.

Note that this doesn't include the manual page and doesn't set the
version you get with `make`.

### Building

If you have Go and make installed, a simple:

```
$ make
```

or, for a Windows executable,

```
$ make tkey-sign.exe
```

should build `tkey-sign`. A pre-compiled signer device app binary is
included in the repo and will be automatically embedded.

Cross compiling the usual Go way with `GOOS` and `GOARCH` environment
variables works for most targets but currently doesn't work for
`GOOS=darwin` since the `go.bug.st/serial` package relies on macOS
shared libraries for port enumeration.

### Building with tkey-builder

If you want to use the tkey-builder image and you have `make` you can
run:

```
$ podman pull ghcr.io/tillitis/tkey-builder:4
$ make podman
```

or run tkey-builder directly with Podman:

```
$ podman run --rm --mount type=bind,source=$(CURDIR),target=/src -w /src -it ghcr.io/tillitis/tkey-builder:4 make -j
```

Note that building with Podman by default creates a Linux binary. Set
`GOOS` and `GOARCH` with `-e` in the call to `podman run` to desired
target. Again, this won't work with a macOS target.

### Installing on Linux

You can install `tkey-sign` and reload the Linux udev rules to get
access to the TKey with:

```
$ sudo make install
$ sudo make reload-rules
```

### Reproducible builds

You should be able to build a binary that is a exact copy of our
release binaries if you use the same Go compiler, at least for the
statically linked Linux and Windows binaries.

Please see [the official
releases](https://github.com/tillitis/tkey-sign-cli/releases) for
digests and details about the build environment.

### Building with another signer

For convenience, and to be able to support `go install` the signer
device app binary is included in `cmd/tkey-sign`.

If you want to replace the signer used you have to:

1. Compile your own signer and place it in `cmd/tkey-sign`.
2. Change the path to the embedded signer in `cmd/tkey-sign/main.go`.
   Look for `go:embed...`.
3. Change the `appName` directly under the `go:embed` to whatever your
   signer is called, so the agent reports this correctly with
   `--version`.
4. Compute a new SHA-512 hash digest for your binary, typically by
   something like `sha512sum cmd/tkey-sign/signer.bin-v0.0.7` and put
   the resulting output in the file `signer.bin.sha512` at the top
   level.
5. `make` in the top level.

## Building the signer

1. See [the Devoloper Handbook](https://dev.tillitis.se/) for setup of
   development tools. We recommend you use tkey-builder.
2. See the instructions in the [tkey-device-signer
   repo](https://github.com/tillitis/tkey-device-signer).
3. Copy its `signer/app.bin` to
   `cmd/tkey-sign/signer.bin-${signer_version}` and run `make`.

To help prevent unpleasant surprises we keep a digest of the signer in
`cmd/tkey-ssh-agent/signer.bin.sha512`. The compilation will fail if
this is not the expected binary. If you really intended to build with
another signer, see [Building with another
signer](#building-with-another-signer) above.

## Licenses and SPDX tags

Unless otherwise noted, the project sources are copyright Tillitis AB,
licensed under the terms and conditions of the "BSD-2-Clause" license.
See [LICENSE](LICENSE) for the full license text.

Until Nov 14, 2025, the license was GPL-2.0 Only.

External source code we have imported are isolated in their own
directories. They may be released under other licenses. This is noted
with a similar `LICENSE` file in every directory containing imported
sources.

The project uses single-line references to Unique License Identifiers
as defined by the Linux Foundation's [SPDX project](https://spdx.org/)
on its own source files, but not necessarily imported files. The line
in each individual source file identifies the license applicable to
that file.

The current set of valid, predefined SPDX identifiers can be found on
the SPDX License List at:

https://spdx.org/licenses/

We attempt to follow the [REUSE
specification](https://reuse.software/).
