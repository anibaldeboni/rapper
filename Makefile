.PHONY: lint test build install_deps mocks

default: all

all: lint test

lint:
	@golangci-lint run -v

test: install_deps
	@go run gotest.tools/gotestsum@latest --format pkgname -- $(go list ./... | grep -v mock) -v -cover -race ./...

build:
	@scripts/build.sh

install_deps:
	@go get -v ./...

mocks:
	@go generate ./...
