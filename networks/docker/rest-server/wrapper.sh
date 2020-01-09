#!/usr/bin/env sh
# This software is Copyright (c) 2019-2020 e-Money A/S. It is not offered under an open source license.
#
# Please contact partners@e-money.com for licensing related questions.


##
## Input parameters
##
BINARY=/emoney/${BINARY:-emcli-linux}
NODE=${NODE:-http://192.168.10.2:80}

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

echo "Connecting to node ${NODE}"
"$BINARY" "$@" --node "$NODE" --trust-node --laddr tcp://0.0.0.0:1317

chmod 777 -R /emoney

