# Binary name
BINARY_NAME=lazyprisma

# Build directory
BUILD_DIR=.

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

.PHONY: all build clean run test install help

all: build

## build: Build the binary
build:
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) .

## clean: Remove build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY_NAME)
	rm -f ./*.tar.gz

## run: Build and run the application
run: build
	./$(BINARY_NAME)

## test: Run tests
test:
	$(GOTEST) -v ./...

## deps: Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

## install: Install the binary to GOPATH/bin
install:
	$(GOCMD) install

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

BINARY_NAME=lazyprisma
VERSION ?= 0.1.7
DIST_DIR=dist

build-all:
	mkdir -p $(DIST_DIR)/darwin-amd64
	mkdir -p $(DIST_DIR)/darwin-arm64
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(DIST_DIR)/darwin-amd64/$(BINARY_NAME) .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(DIST_DIR)/darwin-arm64/$(BINARY_NAME) .

# package: build-all
# 	tar -czvf $(BINARY_NAME)-v$(VERSION)-darwin-amd64.tar.gz -C $(DIST_DIR)/darwin-amd64 $(BINARY_NAME)
# 	tar -czvf $(BINARY_NAME)-v$(VERSION)-darwin-arm64.tar.gz -C $(DIST_DIR)/darwin-arm64 $(BINARY_NAME)

# 	rm -rf $(DIST_DIR)

package: build-all
	tar -czvf $(BINARY_NAME)-v$(VERSION)-darwin-amd64.tar.gz -C $(DIST_DIR)/darwin-amd64 $(BINARY_NAME)
	tar -czvf $(BINARY_NAME)-v$(VERSION)-darwin-arm64.tar.gz -C $(DIST_DIR)/darwin-arm64 $(BINARY_NAME)

	@echo "\n==> SHA256 checksums for v$(VERSION)\n"
	@echo "amd64:"
	@shasum -a 256 $(BINARY_NAME)-v$(VERSION)-darwin-amd64.tar.gz | awk '{print $$1}'
	@echo ""
	@echo "arm64:"
	@shasum -a 256 $(BINARY_NAME)-v$(VERSION)-darwin-arm64.tar.gz | awk '{print $$1}'
	@echo "\n✅ Done.\n"

	rm -rf $(DIST_DIR)
