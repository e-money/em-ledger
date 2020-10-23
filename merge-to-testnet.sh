#!/bin/bash

if [ "$1" == "" ]; then
    echo "Missing emoney-1 export file argument."
    exit 1
fi

make clean local-testnet

sleep 10

docker-compose down

./build/emd export --home build/node0/ > networks/utils/testnet.export.json

./build/emd migrate "$1" > networks/utils/emoney2.export.json

./build/emd unsafe-reset-all --home build/node0/
./build/emd unsafe-reset-all --home build/node1/
./build/emd unsafe-reset-all --home build/node2/
./build/emd unsafe-reset-all --home build/node3/

pushd networks/utils/
python3 transplant-validators.py
popd

cp networks/utils/output.json ./build/node0/config/genesis.json
cp networks/utils/output.json ./build/node1/config/genesis.json
cp networks/utils/output.json ./build/node2/config/genesis.json
cp networks/utils/output.json ./build/node3/config/genesis.json

# Cleanup
rm networks/utils/output.json
rm networks/utils/testnet.export.json
rm networks/utils/emoney2.export.json