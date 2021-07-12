#!/bin/sh

set -e

# copied and modified from https://github.com/informalsystems/ibc-rs/blob/master/scripts/init-hermes

usage() {
  echo "Usage: $0 CONFIG_FILE e-money_ID gaia_ID"
  echo "Example: $0 ./emoney-config.toml emoney gaia"
  exit 1
}

missing() {
  echo "Missing $1 parameter. Please check if all parameters were specified."
  usage
}

if [ -z "$1" ]; then
  missing "CONFIG_FILE"
fi

if [ -z "$2" ]; then
  missing "EMONEY_0_ID"
fi

if [ -z "$3" ]; then
  missing "GAIA_ID"
fi

if [ "$#" -gt 3 ]; then
  echo "Incorrect number of parameters."
  usage
fi

if ! [ -x "$(which hermes)" ]; then
  echo "Error: hermes is not installed: install from https://hermes.informal.systems/installation.html" >&2
  exit 1
fi

CONFIG_FILE="$1"
EMONEY_ID="$2"
GAIA_ID="$3"
EMONEY_DATA="../build"
GAIA_DATA="../build/gdata"
CONFIG_FILE="emoney-config.toml"

if ! [ -f "$CONFIG_FILE" ]; then
  echo "[CONFIG_FILE] ($1) does not exist or is not a file."
  usage
fi

if ! grep -q -s "$EMONEY_0_ID" "$CONFIG_FILE"; then
  echo "error: configuration for chain [$EMONEY_0_ID] does not exist in file $CONFIG_FILE."
  usage
fi

if ! grep -q -s "$GAIA_ID" "$CONFIG_FILE"; then
  echo "error: configuration for chain [$GAIA_ID] does not exist in file $CONFIG_FILE."
  usage
fi

# add the key seeds to the keyring of each chain
echo "Importing keys..."
hermes -c "$CONFIG_FILE" keys add "$EMONEY_ID" -f "./key1.json"
# hermes -- -c "$CONFIG_FILE" keys add "$EMONEY_ID" -f "./auth-key.json"
hermes -c "$CONFIG_FILE" keys add "$GAIA_ID" -f "$GAIA_DATA/$GAIA_ID/user_seed.json"
hermes -c "$CONFIG_FILE" keys add "$GAIA_ID" -f "$GAIA_DATA/$GAIA_ID/user2_seed.json" -n user2

echo "Done!"