# @rosen-bridge/tss

## 3.0.1

### Patch Changes

- fix signPromised return type

## 3.0.0

### Major Changes

- change `sign` function to protected
  change `signPromised` function to abstract
  add `chainCode` and `derivationPath` to `signPromised` arguments and remove them from constructor

## 2.0.1

### Patch Changes

- move common function (randomKey) to abstract class EncryptionHandler

## 2.0.0

### Major Changes

- add chainCode as required to TssSigner class config
- change TssSigner to abstract with abstract `getSignExtraData` and `handleSuccessfulSign` functions

### Minor Changes

- add EcdsaSigner class that implements TssSigner with ECDSA signer
- add EddsaSigner class that implements TssSigner with EdDSA signer
- add `crypto` and `chainCode` parameters to sign request
- add `crypto` parameter to update threshold request
