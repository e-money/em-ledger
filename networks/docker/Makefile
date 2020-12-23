# Makefile for the "emdnode", "avanode" docker images.
GITHUB_DIR= ${GOPATH}/src/github.com
AVALANCHE_SRC_DIR := $(GITHUB_DIR)/ava-labs/avalanchego
AVALANCHE_DEST_DIR := ./avanode/avalanchego/

all: emdnode rest-server

emdnode:
	docker build --tag emoney/emdnode emdnode

rest-server:
	docker build --tag emoney/rest-server rest-server

avalanchefiles:
	@echo "\e[94mEnsure you have cloned avalanchego: \e[92mgit clone https://github.com/ava-labs/avalanchego.git $(GITHUB_DIR)"
	rsync --exclude='.ci/' --exclude='.github/' -at $(AVALANCHE_SRC_DIR) ./avanode/
	rsync -at ./avanode/coreth-config.json $(AVALANCHE_DEST_DIR)

avanode: clean avalanchefiles
	docker build --tag emoney/avanode avanode

avanodeup: avanode
	docker-compose -f avanode/docker-compose.yml up -d

avanodedown:
	docker-compose -f avanode/docker-compose.yml down

clean:
	rm -rf avanode/avalanchego

.PHONY: all rest-server emdnode avanode clean avanalanchefiles avanodeup avanodedown
