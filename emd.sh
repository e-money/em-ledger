#!/bin/bash

# Purpose of this script is to provide easy access to emcli while the executables
# are compiled for linux in order to run in docker.

go run cmd/emd/*.go "$@"
