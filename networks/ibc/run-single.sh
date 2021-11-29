#!/bin/bash

set -xe

./start-single

./ibc-relay.sh

killall -q cosmovisor || echo > /dev/null 2>&1
./gm stop

echo "The relay test completed successfully!"