APP_NAME := glcli
BUILD_DIR := dist
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build run clean install vet

build:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/glcli/

run:
	go run ./cmd/glcli/

clean:
	rm -rf $(BUILD_DIR)

install: build
	cp $(BUILD_DIR)/$(APP_NAME) $(GOPATH)/bin/$(APP_NAME) 2>/dev/null || cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/$(APP_NAME)

vet:
	go vet ./...
