# Release notes

## Upcoming release

- Introduce package signify. Export Signify types to import and export
  them to buffers and files.

- Add support for BLAKE2s hashing.

## v1.1.1

- Update tkeyclient to v1.3.1 to handle TKey Unlocked (product ID 8)
  as a Bellatrix when it comes to USS digest handling.

- Only allow `--force-full-uss` when either `--uss` or `--uss-file` is
  used.

Full
[changelog](https://github.com/tillitis/tkey-sign-cli/compare/v1.1.0...v1.1.1).

## v1.1.0

- Update tkeyclient version because of a vulnerability leaving some
  USSs unused. Keys might have changed since earlier versions! Read
  more here:

  https://github.com/tillitis/tkeyclient/security/advisories/GHSA-4w7r-3222-8h6v

  The error is only triggered if you use `tkey-sign-cli` with the
  `--uss` or `--uss-file` flags and use an affected USS. An affected
  USS hashes to a digest with a 0 (zero) in the first byte.

  Follow these steps to identify if you are affected:
    1. Run `tkey-sign -G -p key.pub --uss`
    2. Type in your USS.
    3. Remove and reinsert the TKey.
    4. Run `tkey-sign -G -p key2.pub`
    5. Compare the `key.pub` and `key2.pub` files. If they have the same
       contents your USS is vulnerable.

  If your USS are affected, you have three options:
    1. Not using a USS and keep your signing keys.
    2. Keep using the USS and get new signing keys.
    3. Use another USS and get new signing keys.

- Add a new option flag: `--force-full-uss` to force full use of the
  32 byte USS digest.

Full
[changelog](https://github.com/tillitis/tkey-sign-cli/compare/v1.0.1...v1.1.0).

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
