# Makefile for AI Image Generator (Golang)

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GORUN=$(GOCMD) run

# Binary name
BINARY_NAME=ai-image-generator
BINARY_UNIX=$(BINARY_NAME)_unix

# Default target
.PHONY: all
all: build

# Build for current platform
.PHONY: build
build:
	$(GOBUILD) -o $(BINARY_NAME) -v

# Run the application
.PHONY: run
run:
	$(GORUN) main.go

# Clean build files
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

# Install dependencies
.PHONY: deps
deps:
	$(GOMOD) tidy

# Setup configuration
.PHONY: setup-config
setup-config:
	@echo "Setting up configuration..."
	@if not exist config.toml (\
		copy config.example.toml config.toml && \
		echo "Configuration file created: config.toml" && \
		echo "Please edit config.toml to set your API keys"\
	) else (\
		echo "Configuration file already exists: config.toml"\
	)

# Validate configuration
.PHONY: check-config
check-config:
	@echo "Validating configuration..."
	@$(GORUN) -tags validate main.go

# Clean configuration (remove config.toml)
.PHONY: clean-config
clean-config:
	@if exist config.toml (\
		del config.toml && \
		echo "Removed config.toml"\
	) else (\
		echo "No config.toml found"\
	)

# Run tests
.PHONY: test
test:
	$(GOTEST) -v ./...

# Build for Windows
.PHONY: build-win
build-win:
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME).exe -v

# Build for Linux
.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v

# Build for macOS
.PHONY: build-mac
build-mac:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v

# Install
.PHONY: install
install:
	$(GOGET) -u github.com/gin-gonic/gin
	$(GOGET) -u github.com/google/uuid
	$(GOGET) -u github.com/sashabaranov/go-openai

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all          - Build for current platform (default)"
	@echo "  build        - Build for current platform"
	@echo "  run          - Run the application"
	@echo "  clean        - Clean build files"
	@echo "  deps         - Install dependencies"
	@echo "  setup-config - Create config.toml from template"
	@echo "  check-config - Validate configuration"
	@echo "  clean-config - Remove config.toml"
	@echo "  test         - Run tests"
	@echo "  build-win    - Build for Windows"
	@echo "  build-linux  - Build for Linux"
	@echo "  build-mac    - Build for macOS"
	@echo "  install      - Install required packages"
	@echo "  help         - Show this help message"