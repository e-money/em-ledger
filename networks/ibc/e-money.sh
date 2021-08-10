#!/bin/bash

set -e

EMONEY_HOME="../.."
EMONEY_DC_YML="$EMONEY_HOME/docker-compose.yml"

# Check platform
platform='unknown'
unamestr=$(uname)
if [ "$unamestr" = 'Linux' ]; then
   platform='linux'
fi

# Set proper defaults and change ports (use a different sed for Mac or Linux)
echo "Stopping the e-money testnet to make IBC appropriate genesis edits..."
set -x
docker-compose -f $EMONEY_DC_YML stop || true
for (( i = 0; i < 4; i++ )); do
  if [ $platform = 'linux' ]; then
    sed -i 's#"172800s"#"200s"#g' $EMONEY_HOME/build/node$i/config/genesis.json
    sed -i 's/timeout_commit = "5s"/timeout_commit = "1s"/g' $EMONEY_HOME/build/node$i/config/config.toml
    sed -i 's/timeout_propose = "3s"/timeout_propose = "1s"/g' $EMONEY_HOME/build/node$i/config/config.toml
    sed -i 's/index_all_keys = false/index_all_keys = true/g' $EMONEY_HOME/build/node$i/config/config.toml
  else
    sed -i '' 's#"172800s"#"200s"#g' $EMONEY_HOME/build/node$i/config/genesis.json
    sed -i '' 's/timeout_commit = "5s"/timeout_commit = "1s"/g' $EMONEY_HOME/build/node$i/config/config.toml
    sed -i '' 's/timeout_propose = "3s"/timeout_propose = "1s"/g' $EMONEY_HOME/build/node$i/config/config.toml
    sed -i '' 's/index_all_keys = false/index_all_keys = true/g' $EMONEY_HOME/build/node$i/config/config.toml
  fi
done
docker-compose -f $EMONEY_DC_YML start
sleep 13