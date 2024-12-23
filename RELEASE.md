# Release notes

## v1.0.1

- Normalize line endings of user input when asked to overwrite a file.
This fixes an issue on Windows where a file was never overwritten
regardless of user input.
- [tkeyutil](https://github.com/tillitis/tkeyutil) has been updated to
v0.0.9. This resolves a bug on USS input for Windows.
- [tkeyclient](https://github.com/tillitis/tkeyclient) has been
updated to v1.1.0.
- [tkeysign](https://github.com/tillitis/tkeysign) has been updated to
v1.0.1.
- Update Go packages.

Full
[changelog](https://github.com/tillitis/tkey-sign-cli/compare/v1.0.0...v1.0.1).

## v1.0.0

- `--version` now also outputs version of embedded device app.
- Builds releases and OS packages with
  [goreleaser](https://goreleaser.com/).
- [tkey-device-signer](https://github.com/tillitis/tkey-device-signer)
  has been updated to v1.0.0. WARNING: Breaks CDI! Generates new key pair.
- [tkeyclient](https://github.com/tillitis/tkeyclient) has been
  updated to v1.0.0.
- [tkeysign](https://github.com/tillitis/tkeysign) has been updated to
  v1.0.0.

Full
[changelog](https://github.com/tillitis/tkey-sign-cli/compare/v0.0.8...v1.0.0).

## v0.0.8

- Using version v0.0.8 of signer.bin from tkey-device-signer.
- Stores signature and public key in OpenBSD signify(1) format.
- Flags work more like signify. Not 100% compatible since we don't
  have '-s' (private key is in TKey, remember?) and we require '-p'
  when signing to verify that the exported public key is the same as
  the TKey's.
- Adds script signify-verify to verify tkey-sign signatures with
  signify.
- Adds functionality to sign arbitrary large files by doing a SHA512
  over the file and signing the digest.

NOTE: the update of the signer.bin version gives the TKey a different
identity (CDI) compare to earlier releases, i.e., the key pair used for
signing and verifying will be different.

Full
[changelog](https://github.com/tillitis/tkey-sign-cli/compare/v0.0.7...v0.0.8).

## v0.0.7

Migrated from https://github.com/tillitis/tillitis-key1-apps/
