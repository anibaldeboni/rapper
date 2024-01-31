.PHONY: lint test build install_deps

default: all

all: lint test

lint:
	@golangci-lint run -v

test: install_deps
	@go run gotest.tools/gotestsum --format pkgname -- -v -cover ./...

build:
	@scripts/build.sh

install_deps:
	@go get -v ./...
