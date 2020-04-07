BINARY := ValueTransformer
GOOS   ?= $(shell go env | grep GOOS | cut -d'"' -f2)

.PHONY: all

all: bin/$(GOOS)/$(BINARY) Makefile

bin/%/$(BINARY):
  GOOS=$* GOARCH=amd64 go build -mod=vendor -v -i -o $(BINARY) $(BINARY).go
  
