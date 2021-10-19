# Makefile for the "emdnode" docker image.

all: emdnode cosmovisor test-upg

emdnode:
	docker build --tag emoney/emdnode emdnode

cosmovisor:
	docker build cosmovisor --tag emoney/cosmovisor

test-upg:
	docker build test-upg --build-arg version=v44-upg-test --tag emoney/test-upg

# todo (reviewer): please note the rest-server is not a command anymore. Please enable `api` section in `app.toml` instead

.PHONY: all emdnode cosmovisor test-upg