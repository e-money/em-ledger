# Makefile for the "emdnode" docker image.

all: emdnode

emdnode:
	docker build --tag emoney/emdnode emdnode

cosmovisor:
	docker build cosmovisor --tag emoney/cosmovisor

# todo (reviewer): please note the rest-server is not a command anymore. Please enable `api` section in `app.toml` instead

.PHONY: all emdnode cosmovisor

