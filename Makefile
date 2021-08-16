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
		  -X github.com/cosmos/cosmos-sdk/version.AppName=emd \
		  -X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
		  -X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)"

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

build:
	go build -mod=readonly $(BUILD_FLAGS) -o build/emd$(BIN_PREFIX) ./cmd/emd

lint:
	golangci-lint run

# go get mvdan.cc/gofumpt
fmt:
	gofumpt -w **/*.go

# go get go get github.com/daixiang0/gci
imp:
	gci -w **/*.go

install:
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/emd

build-linux:
	# Linux images for docker-compose
	# CGO_ENABLED=0 added to solve this issue: https://stackoverflow.com/a/36308464
	BIN_PREFIX=-linux LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

build-all: build-linux build/cosmovisor build/emdupg build/emdupg-linux
	$(MAKE) build

build-docker:
	$(MAKE) -C networks/docker/ all

run-single-node: clean
	go run cmd/daemon/*.go init
	go run cmd/daemon/*.go start

test:
	go test -mod=readonly ./...

bdd-test:
	go test -mod=readonly -v -p 1 -timeout 1h --tags="bdd" bdd_test.go multisigauthority_test.go authority_test.go market_test.go buyback_test.go capacity_test.go staking_test.go bep3swap_test.go upgrade_test.go

github-ci: build-linux
	$(MAKE) test
	$(MAKE) proto-lint

local-testnet:
	go test -mod=readonly -v --tags="bdd" bdd_test.go localnet_test.go

local-testnet-reset:
	./build/emd unsafe-reset-all --home build/node0/
	./build/emd unsafe-reset-all --home build/node1/
	./build/emd unsafe-reset-all --home build/node2/
	./build/emd unsafe-reset-all --home build/node3/

clean:
	rm -rf ./build ./data ./config

##################### upgrade artifacts
cosmovisor:
	go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@latest

build/cosmovisor:
	docker run --rm --entrypoint cat emoney/cosmovisor /go/bin/cosmovisor > build/cosmovisor
	chmod +x build/cosmovisor

build/emdupg:
	docker run --rm --entrypoint cat emoney/test-upg /go/src/em-ledger/build/emd > "build/emdupg"
	chmod +x "build/emdupg"

build/emdupg-linux:
	docker run --rm --entrypoint cat emoney/test-upg /go/src/em-ledger/build/emd-linux > "build/emdupg-linux"
	chmod +x "build/emdupg-linux"

license:
	GO111MODULE=off go get github.com/google/addlicense/
	addlicense -f LICENSE .

.PHONY: build build-linux cosmovisor clean test bdd-test build-docker license

###############################################################################
###                                Protobuf                                 ###
###############################################################################
DOCKER := $(shell which docker)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf

proto-all: proto-format proto-lint proto-gen proto-swagger-gen

proto-gen:
	@echo "Generating Protobuf files"
	$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace tendermintdev/sdk-proto-gen sh ./scripts/protocgen.sh

proto-format:
	@echo "Formatting Protobuf files"
	$(DOCKER) run --rm -v $(CURDIR):/workspace \
	--workdir /workspace tendermintdev/docker-build-proto \
	find ./ -not -path "./third_party/*" -name *.proto -exec clang-format -i {} \;

proto-swagger-gen:
	@./scripts/protoc-swagger-gen.sh

proto-lint:
	@$(DOCKER_BUF) lint --error-format=json

proto-check-breaking:
	@$(DOCKER_BUF) breaking --against-input $(HTTPS_GIT)#branch=master

.PHONY: proto-all proto-gen proto-swagger-gen proto-format proto-lint proto-check-breaking