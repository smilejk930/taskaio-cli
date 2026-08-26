BINARY_NAME := taskaio
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -s -w -X github.com/taskaio/taskaio-cli/cmd.Version=$(VERSION) -X github.com/taskaio/taskaio-cli/cmd.Commit=$(COMMIT) -X github.com/taskaio/taskaio-cli/cmd.Date=$(DATE)
INSTALL_DIR ?= $(HOME)/.local/bin

.PHONY: all build build-all test lint clean install

all: build

build:
	@mkdir -p bin
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) .

build-all:
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME)-linux-amd64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME)-linux-arm64 .

test:
	go test -v ./...

lint:
	go vet ./...

install: build
	@mkdir -p $(INSTALL_DIR)
	install -m 0755 bin/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed taskaio to $(INSTALL_DIR)/$(BINARY_NAME)"

clean:
	rm -rf bin dist
