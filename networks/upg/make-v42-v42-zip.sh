#!/bin/zsh

# ensure ../../build/emd is the binary you want to generate the intended sdk
# genesis versioned config files and db and build args i.e. fast-consensus

set -e

rm -rf .emd

pushd ../../
make clean
make build-fast-consensus
popd

# start emd
./start-full-cv &

# generate ~2 blocks
sleep 5

EMD=$PWD/.emd/cosmovisor/current/bin/emd

$EMD tx authority schedule-upgrade authoritykey test-upg-0.2.0 --upgrade-height 8 --yes --from authoritykey --home=".emd" --node tcp://localhost:26657 --chain-id localnet_reuse --keyring-backend test

sleep 2

# stop emd to back it up
killall -q cosmovisor || echo > /dev/null 2>&1
killall -q emd || echo > /dev/null 2>&1
sleep 5

ZIP_FILE=/tmp/v42-v42.zip

zip "$ZIP_FILE" -r .emd

# delete binary files from the zip
zip --delete "$ZIP_FILE" ".emd/cosmovisor/current/*"

printf "Created %s\n" "$ZIP_FILE"
