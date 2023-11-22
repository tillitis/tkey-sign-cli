# Release notes

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
signing and verifiying will be different.

Full
[changelog](https://github.com/tillitis/tkey-device-signer/compare/v0.0.7...v0.0.8).

## v0.0.7

Migrated from https://github.com/tillitis/tillitis-key1-apps/
