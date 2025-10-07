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
