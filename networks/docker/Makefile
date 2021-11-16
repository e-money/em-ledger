# Makefile for the "em-ledger" docker images.

all: emdnode cosmovisor test-upg

# force rebuild when docker is caching images
# use case: switching cosmos-sdk versions
all-f: emdnode-f cosmovisor-f test-upg-f

emdnode:
	docker build --tag emoney/emdnode emdnode

emdnode-f:
	docker build --tag emoney/emdnode emdnode --no-cache

cosmovisor:
	docker build cosmovisor --tag emoney/cosmovisor

cosmovisor-f:
	docker build cosmovisor --tag emoney/cosmovisor --no-cache

test-upg:
	docker build test-upg --build-arg branch=$$(git rev-parse --abbrev-ref HEAD) --build-arg version=test-upg-0.2.0 --tag emoney/test-upg --no-cache

test-upg-f:
	docker build test-upg --build-arg branch=$$(git rev-parse --abbrev-ref HEAD) --build-arg version=test-upg-0.2.0 --tag emoney/test-upg --no-cache

# todo (reviewer): please note the rest-server is not a command anymore. Please enable `api` section in `app.toml` instead

.PHONY: all emdnode cosmovisor test-upg all-f emdnode-f cosmovisor-f test-upg-f
