
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

## Building

You have two options, either our OCI image
`ghcr.io/tillitis/tkey-builder` for use with a rootless podman setup,
or native tools.

With podman you should be able to use:

```
$ ./build-podman.sh
```

which requires at least `git` and `make` besides podman. See below for
setting things up.

With native tools you should be able to use our build script:

```
$ ./build.sh
```

Both of these also clones and builds the [TKey device
libraries](https://github.com/tillitis/tkey-libs) and the [signer
device app](https://github.com/tillitis/tkey-device-signer) first.

If you want to do it manually please inspect the build script, but
basically you clone the `tkey-libs` and `tkey-device-signer` repos,
build the signer, copy it's `app.bin` to `cmd/tkey-sign/signer.bin`
and run `make`.

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
