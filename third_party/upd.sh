#!/bin/zsh

set -e

SDK_SRC=$(go list -f '{{.Dir}}' -m github.com/cosmos/cosmos-sdk)
cp -r $SDK_SRC/proto/cosmos/* ./proto/cosmos/
cp -r $SDK_SRC/third_party/proto/* ./proto/

#ibc-go repo
SDK_SRC=$(go list -f '{{.Dir}}' -m github.com/cosmos/ibc-go)
cp -r $SDK_SRC/proto/ibc/* ./proto/ibc/

ls **/*.proto | xargs chmod 666