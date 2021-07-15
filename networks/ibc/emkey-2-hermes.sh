#!/bin/sh

set -e

usage() {
  echo "Usage: $0 CONFIG_FILE e-money-chain-id"
  echo "Example: $0 ./emoney-config.toml localnet_reuse"
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
  missing "EMONEY_ID"
fi

if [ "$#" -gt 2 ]; then
  echo "Incorrect number of parameters."
  usage
fi

if ! [ -x "$(which hermes)" ]; then
  echo "Error: hermes is not installed: install from https://hermes.informal.systems/installation.html" >&2
  exit 1
fi

CONFIG_FILE="$1"
EMONEY_ID="$2"

if ! [ -f "$CONFIG_FILE" ]; then
  echo "[CONFIG_FILE] ($1) does not exist or is not a file."
  usage
fi

if ! grep -q -s "$EMONEY_ID" "$CONFIG_FILE"; then
  echo "error: configuration for chain [$EMONEY_ID] does not exist in file $CONFIG_FILE."
  usage
fi

# add the key seeds to the keyring of each chain
echo "Importing keys..."
hermes -c "$CONFIG_FILE" keys add "$EMONEY_ID" -f "./em-key1.json"