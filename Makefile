# Vibe CLI Makefile

# Binary name
BINARY=vibe

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOFMT=gofmt

# Version info (can be overridden)
VERSION?=dev
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Linker flags for version info
LDFLAGS=-ldflags "-X github.com/user/vibe/cmd.Version=$(VERSION) -X github.com/user/vibe/cmd.GitCommit=$(GIT_COMMIT) -X github.com/user/vibe/cmd.BuildTime=$(BUILD_TIME)"

# Install location
INSTALL_PATH=/usr/local/bin

.PHONY: all build install uninstall test fmt clean help

## build: Build the binary (default)
build:
	$(GOBUILD) -o $(BINARY) .

## build-release: Build with version info and optimizations
build-release:
	$(GOBUILD) $(LDFLAGS) -ldflags="-s -w" -o $(BINARY) .

## install: Build and install to /usr/local/bin
install: build
	sudo cp $(BINARY) $(INSTALL_PATH)/$(BINARY)
	@echo "Installed $(BINARY) to $(INSTALL_PATH)"

## link: Create symlink to current binary (for development)
link: build
	sudo ln -sf $(PWD)/$(BINARY) $(INSTALL_PATH)/$(BINARY)
	@echo "Created symlink: $(INSTALL_PATH)/$(BINARY) -> $(PWD)/$(BINARY)"

## uninstall: Remove from /usr/local/bin
uninstall:
	sudo rm -f $(INSTALL_PATH)/$(BINARY)
	@echo "Removed $(BINARY) from $(INSTALL_PATH)"

## test: Run tests
test:
	$(GOTEST) ./...

## test-verbose: Run tests with verbose output
test-verbose:
	$(GOTEST) -v ./...

## fmt: Format code
fmt:
	$(GOFMT) -w .

## tidy: Tidy go modules
tidy:
	$(GOMOD) tidy

## clean: Remove binary
clean:
	rm -f $(BINARY)

## help: Show this help
help:
	@echo "Vibe CLI - Available targets:"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed 's/^/ /'
	@echo ""
	@echo "Examples:"
	@echo "  make              # Build the binary"
	@echo "  make install      # Build and install to PATH"
	@echo "  make link         # Create symlink for development"
	@echo "  make test         # Run tests"

# Default target
all: build
