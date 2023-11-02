# Release notes

## v0.0.8

- Stores signature and public key in OpenBSD signify(1) format.
- Flags work more like signify. Not 100% compatible since we don't
  have '-s' (private key is in TKey, remember?) and we require '-p'
  when signing to verify that the exported public key is the same as
  the TKey's.
- Adds script signify-verify to verify tkey-sign signatures with
  signify.
- Adds functionality to sign arbitrary large files by doing a SHA512
  over the file and signing the digest.

## v0.0.7

Migrated from https://github.com/tillitis/tillitis-key1-apps/
