#!/bin/bash

set -xe

./start

./ibc-relay.sh

./gm stop

echo "The relay test completed successfully!"