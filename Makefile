export GO111MODULE=on

VERSION = $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT  = $(shell git log -1 --format='%H')
LEDGER_ENABLED ?= true

# process build tags
build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags
ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=e-money \
          -X github.com/cosmos/cosmos-sdk/version.ServerName=emd \
          -X github.com/cosmos/cosmos-sdk/version.ClientName=emcli \
          -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
          -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
          -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)"

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

install:
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/emd
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/emcli

build:
	go build -mod=readonly $(BUILD_FLAGS) -o build/emd$(BIN_PREFIX) ./cmd/emd
	go build -mod=readonly $(BUILD_FLAGS) -o build/emcli$(BIN_PREFIX) ./cmd/emcli

build-linux:
	# Linux images for docker-compose
	# CGO_ENABLED=0 added to solve this issue: https://stackoverflow.com/a/36308464
	BIN_PREFIX=-linux LEDGER_ENABLED=false GOOS=linux CGO_ENABLED=0 GOARCH=amd64 $(MAKE) build

build-all: build-linux
	$(MAKE) build

build-docker:
	$(MAKE) -C networks/docker/ all

run-single-node: clean
	go run cmd/daemon/*.go init
	go run cmd/daemon/*.go start

test:
	go test -mod=readonly ./...

bdd-test:
	go test -mod=readonly -v -p 1 --tags="bdd" bdd_test.go staking_test.go restricted_denom_test.go multisigauthority_test.go authority_test.go capacity_test.go market_test.go buyback_test.go

local-testnet:
	go test -mod=readonly -v --tags="bdd" bdd_test.go localnet_test.go

clean:
	rm -rf ./build ./data ./config

license:
	GO111MODULE=off go get github.com/google/addlicense/
	addlicense -f LICENSE .


.PHONY: build build-linux clean test bdd-test build-docker license