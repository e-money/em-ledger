#!/bin/bash

set -xe

./start.sh

./ibc-conn.sh
./ibc-gaia-em-relay.sh

# restart chains to simulate upgrade
echo 'Restarting Gaia'
./gm stop
./gm start &

echo 'Restarting em-ledger'
EMONEY_HOME="../.."
EMONEY_DC_YML="$EMONEY_HOME/docker-compose.yml"
docker-compose -f $EMONEY_DC_YML stop || true
docker-compose -f $EMONEY_DC_YML start
sleep 13

# Continue relays
./ibc-em-gaia-relay.sh
./gm stop

echo "$(tput setaf 6)The relay test completed successfully!$(tput sgr 0)"