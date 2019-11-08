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
	go install $(BUILD_FLAGS) ./cmd/emd
	go install $(BUILD_FLAGS) ./cmd/emcli

build:
	go build $(BUILD_FLAGS) -o build/emd$(BIN_PREFIX) ./cmd/emd
	go build $(BUILD_FLAGS) -o build/emcli$(BIN_PREFIX) ./cmd/emcli

build-linux:
	# Linux images for docker-compose
	BIN_PREFIX=-linux LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

build-all: build-linux
	$(MAKE) build

build-docker:
	$(MAKE) -C networks/docker/ all

run-single-node: clean
	go run cmd/daemon/*.go init
	go run cmd/daemon/*.go start

test:
	go test ./...

bdd-test:
	go test -v -p 1 --tags="bdd" bdd_test.go staking_test.go authority_test.go capacity_test.go

clean:
	rm -rf ./build ./data ./config

.PHONY: build build-linux clean test bdd-test build-docker