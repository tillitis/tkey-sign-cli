#! /bin/sh

set -e

if [ -d ../tkey-libs ]
then
    (cd ../tkey-libs; git checkout v0.0.1)
else
    git clone -b v0.0.1 https://github.com/tillitis/tkey-libs.git ../tkey-libs
fi

if [ -d ../tkey-device-signer ]
then
    (cd ../tkey-device-signer; git checkout v0.0.7)
else
    git clone -b v0.0.7 https://github.com/tillitis/tkey-device-signer.git ../tkey-device-signer
fi

make -C ../tkey-libs podman
make -C ../tkey-device-signer podman

cp ../tkey-device-signer/signer/app.bin signer.bin

make podman
