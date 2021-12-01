#!/bin/bash

set -xe

./start-single

./ibc-relay.sh

killall -q cosmovisor || echo > /dev/null 2>&1
./gm stop

echo "$(tput setaf 6)The relay test completed successfully!$(tput sgr 0)"
