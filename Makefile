BINARY_NAME=asactl
VERSION?=0.1.0-dev
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS=-ldflags "-X github.com/emredurukan/asactl/internal/version.Version=$(VERSION) \
                  -X github.com/emredurukan/asactl/internal/version.GitCommit=$(COMMIT) \
                  -X github.com/emredurukan/asactl/internal/version.BuildDate=$(DATE)"

.PHONY: all build test clean install

all: build

build:
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) main.go
	@echo "Built bin/$(BINARY_NAME)"

test:
	go test -v -race ./...

install:
	go install $(LDFLAGS)
	@echo "Installed $(BINARY_NAME) to $(shell go env GOPATH)/bin"

clean:
	rm -rf bin/
