BINARY := ImageRegistryTransformer
GOOS   ?= $(shell go env | grep GOOS | cut -d'"' -f2)

.PHONY: all
all: bin/$(GOOS)/$(BINARY)

bin/$(GOOS)/$(BINARY):
	GOOS=$* GOARCH=amd64 go build -mod=vendor -v -o $(BINARY) ./...

KUSTOMIZE_PLUGIN_HOME=$(realpath $(CURDIR)/../..)
test:
	echo $(KUSTOMIZE_PLUGIN_HOME)
	go test -v -timeout 5m ./...
