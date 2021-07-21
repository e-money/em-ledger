#!/bin/bash

set -e

echo alice balance:
# shellcheck disable=SC2046
emibctokensd --node tcp://localhost:26557 query bank balances $(emibctokensd --home .node keys --keyring-backend="test" show alice -a)

echo bob balance:
# shellcheck disable=SC2046
emibctokensd --node tcp://localhost:26557 query bank balances $(emibctokensd --home .node keys --keyring-backend="test" show bob -a)