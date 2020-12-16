# Makefile for the "emdnode" docker image.

all: emdnode rest-server

emdnode:
	docker build --tag emoney/emdnode emdnode

rest-server:
	docker build --tag emoney/rest-server rest-server

avanode:
	rsync --exclude='.ci/' --exclude='.github/' -at ../../../../ava-labs/avalanchego ./avanode/
	docker build --tag emoney/avanode avanode
	rm -rf avanode/avalanchego

.PHONY: all rest-server emdnode avanode
