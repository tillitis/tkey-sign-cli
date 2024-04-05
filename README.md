
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

Get a public key, possibly modifying the key pair by using a User
Supplied Secret, and storing the public key in file `-p pubkey`.

```
tkey-sign -G/--getkey [-d/--port device] [-s/--speed speed]
[--uss] [--uss-file secret] -p/--public pubkey
```

Sign a file, specified with `-m message` , possibly modifying the
measured key pair by using a User Supplied Secret, and storing the
signature in `-x sigfile` or, by default, in `message.sig`. You need
to supply the public key file as well which `tkey-sign` will
automatically verify that it's the expected public key.

```
tkey-sign -S/--sign [-d/--port device] [-s speed] -m message
[--uss] [--uss-file] -p/--public pubkey [-x sig-file]
```

Verify a signature of file `-m message` with public key in `-p pubkey`.
Signature is by default in `message.sig` but can be specified
with `-x sigfile`. Doesn't need a connected TKey.

```
tkey-sign -V/--verify -m message -p/--public pubkey [-x sigfile]
```

Alternatively you can use OpenBSD's *signify(1)* to verify the
signature but you need to compute the SHA-512 of the file first and
feed that to the verification. We provide a handy script that does
this:

```
signify-verify message pubkey
```

Exit code is 0 on success and non-zero on failure.

See the manual page for details.

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

Note that this doesn't set the version and other stuff you get if you
use `make`.

### Reproducible builds

We're currently building release builds with
[goreleaser](https://goreleaser.com/) using Go 1.19.13.

You should be able to build a binary that is a bit exact reproduction
of our release binaries if you use the same Go compiler, at least for
the statically linked Linux and Windows binaries.

On macOS `tkey-sign` is unfortunately not statically linked. The
binary was built on macOS with uname:

```
Darwin Kernel Version 22.6.0: Wed Jul  5 22:21:53 PDT 2023;
root:xnu-8796.141.3~6/RELEASE_ARM64_T6020 arm64
```

### Build everything

If you want to build it all, including the signer device app, you have
two options, either our OCI image `ghcr.io/tillitis/tkey-builder` for
use with a rootless podman setup, or native tools.

With podman you should be able to use:

```
$ ./build-podman.sh
```

which requires at least `git` and `make` besides podman. See below for
setting things up.

Please note that the Go version in the OCI image is temporarily not
the same as the release process!

With native tools you should be able to use our build script:

```
$ ./build.sh
```

Both of these scripts also clones and builds the [TKey device
libraries](https://github.com/tillitis/tkey-libs) and the [signer
device app](https://github.com/tillitis/tkey-device-signer) first.

If you want to do it manually please inspect the build script, but
basically you clone the `tkey-libs` and `tkey-device-signer` repos,
build the signer, copy it's `app.bin` to
`cmd/tkey-sign/signer.bin-${signer_version}` and run `make`.

You can install `tkey-sign` and reload the udev rules to get access to
the TKey with:

```
$ sudo make install
$ sudo make reload-rules
```

### Installing Podman

On Ubuntu 22.10, running

```
apt install podman rootlesskit slirp4netns
```

should be enough to get you a working Podman setup.

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

If you want to use the `build.sh` and `build-podman.sh` scripts you
have to change the `signer_version` variable and the URL used to clone
the signer device app repo.

## Licenses and SPDX tags

Unless otherwise noted, the project sources are licensed under the
terms and conditions of the "GNU General Public License v2.0 only":

> Copyright Tillitis AB.
>
> These programs are free software: you can redistribute it and/or
> modify it under the terms of the GNU General Public License as
> published by the Free Software Foundation, version 2 only.
>
> These programs are distributed in the hope that it will be useful,
> but WITHOUT ANY WARRANTY; without even the implied warranty of
> MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
> General Public License for more details.

> You should have received a copy of the GNU General Public License
> along with this program. If not, see:
>
> https://www.gnu.org/licenses

See [LICENSE](LICENSE) for the full GPLv2-only license text.

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
