export GO111MODULE=on

VERSION = $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT  = $(shell git log -1 --format='%H')

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=e-money \
          -X github.com/cosmos/cosmos-sdk/version.ServerName=emd \
          -X github.com/cosmos/cosmos-sdk/version.ClientName=emcli \
          -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
          -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
          -X "github.com/cosmos/cosmos-sdk/version.BuildTags="

BUILD_FLAGS = -ldflags '$(ldflags)'

build:
	go build $(BUILD_FLAGS) -o build/emd ./cmd/daemon
	go build $(BUILD_FLAGS) -o build/emcli ./cmd/cli

build-linux:
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

run-single-node: clean
	go run cmd/daemon/*.go init
	go run cmd/daemon/*.go start

clean:
	rm -rf ./build ./data ./config

.PHONY: build build-linux