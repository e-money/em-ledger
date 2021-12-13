#!/usr/bin/make -f

DOCKER := $(shell which docker)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf
HTTPS_GIT := https://github.com/e-money/em-ledger.git
PROJECT_NAME = $(shell git remote get-url origin | xargs basename -s .git)

export GO111MODULE=on

BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT  := $(shell git log -1 --format='%H')

# don't override user values
ifeq (,$(VERSION))
  VERSION := $(shell git describe --exact-match 2>/dev/null)
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH)-$(COMMIT)
  endif
endif

LEDGER_ENABLED ?= true
SDK_PACK := $(shell go list -m github.com/cosmos/cosmos-sdk | sed  's/ /\@/g')
TM_VERSION := $(shell go list -m github.com/tendermint/tendermint | sed 's:.* ::') # grab everything after the space
DOCKER := $(shell which docker)
SDK_PACK := $(shell go list -m github.com/cosmos/cosmos-sdk | sed  's/ /\@/g')
BUILDDIR ?= $(CURDIR)/build
FAST_CONSENSUS ?= false

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

ifeq (cleveldb,$(findstring cleveldb,$(EM_BUILD_OPTIONS)))
  build_tags += gcc cleveldb
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
		  -X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)" \
		  -X github.com/tendermint/tendermint/version.TMCoreSemVer=$(TM_VERSION)

ifeq ($(FAST_CONSENSUS),true)
	ldflags += -X github.com/e-money/em-ledger/cmd/emd/cmd.CreateEmptyBlocksInterval=2s
endif

ifeq (cleveldb,$(findstring cleveldb,$(EM_BUILD_OPTIONS)))
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif
ifeq (,$(findstring nostrip,$(EM_BUILD_OPTIONS)))
  ldflags += -w -s
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'
# check for nostrip option
ifeq (,$(findstring nostrip,$(EM_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
endif

#Print flags when needed
#$(info $$BUILD_FLAGS -> [$(BUILD_FLAGS)])
#$(info )
#$(info $$ldflags -> [$(ldflags)])
#$(info )
#$(info $$EM_BUILD_OPTIONS -> [$(EM_BUILD_OPTIONS)])
#$(info )

build:
	go build -mod=readonly $(BUILD_FLAGS) -o build/emd$(BIN_PREFIX) ./cmd/emd

cosmovisor:
	go install github.com/cosmos/cosmos-sdk/cosmovisor/cmd/cosmovisor@v1.0.0

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

build-reproducible: go.sum
	$(DOCKER) pull tendermintdev/rbuilder:latest
	$(DOCKER) rm latest-build || true
	$(DOCKER) run --volume=$(CURDIR):/sources:ro \
        --env TARGET_PLATFORMS='linux/amd64 darwin/amd64 linux/arm64' \
        --env APP=emd \
        --env VERSION=$(VERSION) \
        --env COMMIT=$(COMMIT) \
        --env LEDGER_ENABLED=$(LEDGER_ENABLED) \
        --name latest-build tendermintdev/rbuilder:latest
	$(DOCKER) cp -a latest-build:/home/builder/artifacts/ $(CURDIR)/

build-linux:
	# Linux images for docker-compose
	# CGO_ENABLED=0 added to solve this issue: https://stackoverflow.com/a/36308464
	BIN_PREFIX=-linux LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

build-all: build-linux build/cosmovisor build/emdupg build/emdupg-linux build/emdupg44 build/emdupg44-linux
	$(MAKE) build

build-fast-consensus:
	FAST_CONSENSUS=true $(MAKE) build-all

build-docker:
	$(MAKE) -C networks/docker/ all

build-docker-f:
	$(MAKE) -C networks/docker/ all-f

run-single-node: clean
	go run cmd/daemon/*.go init
	go run cmd/daemon/*.go start

test:
	go test -mod=readonly ./...

bdd-test:
	go test -mod=readonly -v -p 1 -timeout 1h --tags="bdd" bdd_test.go multisigauthority_test.go authority_test.go market_test.go buyback_test.go capacity_test.go staking_test.go upgrade_test.go

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
	rm -rf $(BUILDDIR)/ artifacts/

##################### upgrade artifacts
build/cosmovisor:
	docker run --rm --entrypoint cat emoney/cosmovisor /go/bin/cosmovisor > build/cosmovisor
	chmod +x build/cosmovisor

build/emdupg:
	docker run --rm --entrypoint cat emoney/test-upg /go/src/em-ledger/build/emd > "build/emdupg"
	chmod +x "build/emdupg"

build/emdupg-linux:
	docker run --rm --entrypoint cat emoney/test-upg /go/src/em-ledger/build/emd-linux > "build/emdupg-linux"
	chmod +x "build/emdupg-linux"

build/emdupg44:
	docker run --rm --entrypoint cat emoney/test-v44 /go/src/em-ledger/build/emd > "build/emdupg44"
	chmod +x "build/emdupg44"

build/emdupg44-linux:
	docker run --rm --entrypoint cat emoney/test-v44 /go/src/em-ledger/build/emd-linux > "build/emdupg44-linux"
	chmod +x "build/emdupg44-linux"

license:
	GO111MODULE=off go get github.com/google/addlicense/
	addlicense -f LICENSE .

.PHONY: build build-linux cosmovisor clean test bdd-test build-docker license

###############################################################################
###                                Protobuf                                 ###
###############################################################################

protoVer=v0.2
protoImageName=tendermintdev/sdk-proto-gen:$(protoVer)
containerProtoGen=$(PROJECT_NAME)-proto-gen-$(protoVer)
containerProtoGenAny=$(PROJECT_NAME)-proto-gen-any-$(protoVer)
containerProtoGenSwagger=$(PROJECT_NAME)-proto-gen-swagger-$(protoVer)
containerProtoFmt=$(PROJECT_NAME)-proto-fmt-$(protoVer)

proto-all: proto-format proto-lint proto-gen proto-swagger-gen

proto-gen:
	@echo "Generating Protobuf files"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoGen}$$"; then docker start -a $(containerProtoGen); else docker run --name $(containerProtoGen) -v $(CURDIR):/workspace --workdir /workspace $(protoImageName) \
		sh ./scripts/protocgen.sh; fi

proto-format:
	@echo "Formatting Protobuf files"
	$(DOCKER) run --rm -v $(CURDIR):/workspace \
	--workdir /workspace tendermintdev/docker-build-proto \
	find ./ -not -path "./third_party/*" -name *.proto -exec clang-format -i {} \;

proto-swagger-gen:
	@echo "Generating Protobuf Swagger"
	@if docker ps -a --format '{{.Names}}' | grep -Eq "^${containerProtoGenSwagger}$$"; then docker start -a $(containerProtoGenSwagger); else docker run --name $(containerProtoGenSwagger) -v $(CURDIR):/workspace --workdir /workspace $(protoImageName) \
		sh ./scripts/protoc-swagger-gen.sh; fi

proto-lint:
	@$(DOCKER_BUF) lint --error-format=json

proto-check-breaking:
	@$(DOCKER_BUF) breaking --against-input $(HTTPS_GIT)#branch=master

.PHONY: proto-all proto-gen proto-swagger-gen proto-format proto-lint proto-check-breaking build-fast-consensus