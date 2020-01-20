# Makefile for the "emdnode" docker image.

all: emdnode rest-server

emdnode:
	docker build --tag emoney/emdnode emdnode

rest-server:
	docker build --tag emoney/rest-server rest-server


.PHONY: all rest-server emdnode

