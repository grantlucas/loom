BINARY := loom
PKG := ./cmd/loom
BUILD_DIR := bin
VERSION ?= dev
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: build test lint clean run

build:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) $(PKG)

test:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BUILD_DIR)

run: build
	$(BUILD_DIR)/$(BINARY)
