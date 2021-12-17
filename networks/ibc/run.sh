#!/bin/bash

set -xe

./start

./ibc-relay.sh

./gm stop

echo "$(tput setaf 6)The relay test completed successfully!$(tput sgr 0)"