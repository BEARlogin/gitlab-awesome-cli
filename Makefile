APP_NAME := glcli
MCP_NAME := glcli-mcp
BUILD_DIR := dist
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build build-mcp build-all run clean install install-mcp vet test lint

build:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/glcli/

build-mcp:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(MCP_NAME) ./cmd/glcli-mcp/

build-all: build build-mcp

run:
	go run ./cmd/glcli/

clean:
	rm -rf $(BUILD_DIR)

install: build
	cp $(BUILD_DIR)/$(APP_NAME) $(GOPATH)/bin/$(APP_NAME) 2>/dev/null || cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/$(APP_NAME)

install-mcp: build-mcp
	cp $(BUILD_DIR)/$(MCP_NAME) $(GOPATH)/bin/$(MCP_NAME) 2>/dev/null || cp $(BUILD_DIR)/$(MCP_NAME) /usr/local/bin/$(MCP_NAME)

vet:
	go vet ./...

test:
	go test -race -count=1 ./...

lint:
	@which golangci-lint > /dev/null 2>&1 || (echo "golangci-lint not found. Install: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...
