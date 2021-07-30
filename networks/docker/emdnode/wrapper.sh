#!/usr/bin/env sh
# This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
#
# Please contact partners@e-money.com for licensing related questions.


##
## Input parameters
##
BINARY=/emoney/${BINARY:-emd}
ID=${ID:-0}
LOG=${LOG:-emd.log}
##
## Assert linux binary
##
if ! [ -f "${BINARY}" ]; then
	echo "The binary $(basename "${BINARY}") cannot be found. Please add the binary to the shared folder. Please use the BINARY environment variable if the name of the binary is not 'emd' E.g.: -e BINARY=emd_my_test_version"
	exit 1
fi
BINARY_CHECK="$(file "$BINARY" | grep 'ELF 64-bit LSB executable, x86-64')"
if [ -z "${BINARY_CHECK}" ]; then
	echo "Binary needs to be OS linux, ARCH amd64"
	exit 1
fi

##
## Run binary with all parameters
##
export EMDHOME="/emoney/node${ID}"

# setup cosmovisor
mkdir -p "$EMDHOME/cosmovisor/genesis/bin"
cp /emoney/emd "$EMDHOME/cosmovisor/genesis/bin"
mkdir -p "$EMDHOME/cosmovisor/upgrades"

export DAEMON_ALLOW_DOWNLOAD_BINARIES=true
export DAEMON_RESTART_AFTER_UPGRADE=true
export DAEMON_HOME=$EMDHOME
# link chain launcher to cosmovisor with linux emd binary
export DAEMON_NAME=emd
BINARY=/emoney/cosmovisor

if [ -d "$(dirname "${EMDHOME}"/"${LOG}")" ]; then
  "$BINARY" --home "$EMDHOME" "$@" | tee "${EMDHOME}/${LOG}"
else
  "$BINARY" --home "$EMDHOME" "$@"
fi

chmod 777 -R /emoney