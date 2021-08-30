#!/bin/sh

set -e

SDK_SRC=$(go list -f '{{.Dir}}' -m github.com/cosmos/cosmos-sdk)

cp -r $SDK_SRC/proto/cosmos ./third_party/proto
cp -r $SDK_SRC/proto/ibc ./third_party/proto
cp -r $SDK_SRC/third_party/* ./third_party
