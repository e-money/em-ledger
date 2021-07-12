#!/bin/bash

set -e

EMONEY_HOME=".."

# Check platform
platform='unknown'
unamestr=`uname`
if [ "$unamestr" = 'Linux' ]; then
   platform='linux'
fi

#for (( i = 0; i < 4; i++ )); do
#  echo $EMONEY_HOME/build/node"$i"
#  if [ $platform = 'linux' ]; then
#  	[ -f $EMONEY_HOME/build/node"$i" ] &&
#  	echo "$EMONEY_HOME/build/node$i exists"
#  else
#    echo "macos"
#  fi
#done

# Set proper defaults and change ports (use a different sed for Mac or Linux)
echo "Stopping the e-money testnet to make IBC appropriate genesis edits..."
set -x
docker-compose -f $EMONEY_HOME/docker-compose.yml stop || true
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
docker-compose -f $EMONEY_HOME/docker-compose.yml start