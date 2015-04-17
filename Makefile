export GOPATH := $(PWD):$(PWD)/gopath
export PATH := $(GOPATH)/bin:$(PATH)
export GOMAXPROCS := 1

GOCMD := go
CAPNP_VERSION := 0.5.1.2
CAPNP_NAME := capnproto-c++-$(CAPNP_VERSION)
CAPNP_CMD := capnp


.PHONY: clean
clean: 
	$(RM) -rf hello

.PHONY: compile
compile: clean 
	${GOCMD} build

DEFAULT_GOAL := compile
