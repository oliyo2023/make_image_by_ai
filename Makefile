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

# Docker parameters
DOCKER_IMAGE_NAME=ai-image-generator
DOCKER_TAG=latest
DOCKER_REGISTRY=

# Docker build
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_TAG) .

# Docker run
.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	docker run -d --name ai-image-generator -p 8000:8000 \
		-v $$(pwd)/public/static/images:/app/public/static/images \
		-v $$(pwd)/logs:/app/logs \
		$(DOCKER_IMAGE_NAME):$(DOCKER_TAG)

# Docker stop
.PHONY: docker-stop
docker-stop:
	@echo "Stopping Docker container..."
	docker stop ai-image-generator || true
	docker rm ai-image-generator || true

# Docker logs
.PHONY: docker-logs
docker-logs:
	@echo "Showing Docker container logs..."
	docker logs -f ai-image-generator

# Docker compose up
.PHONY: docker-compose-up
docker-compose-up:
	@echo "Starting services with docker-compose..."
	docker-compose up -d

# Docker compose down
.PHONY: docker-compose-down
docker-compose-down:
	@echo "Stopping services with docker-compose..."
	docker-compose down

# Docker compose logs
.PHONY: docker-compose-logs
docker-compose-logs:
	@echo "Showing docker-compose logs..."
	docker-compose logs -f

# Setup Docker environment
.PHONY: setup-docker-env
setup-docker-env:
	@echo "Setting up Docker environment..."
	@if not exist .env (\
		copy .env.example .env && \
		echo "Environment file created: .env" && \
		echo "Please edit .env to set your API keys"\
	) else (\
		echo "Environment file already exists: .env"\
	)
	@if not exist public\static\images mkdir public\static\images
	@if not exist logs mkdir logs

# Clean Docker
.PHONY: docker-clean
docker-clean:
	@echo "Cleaning Docker images and containers..."
	docker system prune -f
	docker image prune -f

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all                  - Build for current platform (default)"
	@echo "  build                - Build for current platform"
	@echo "  run                  - Run the application"
	@echo "  clean                - Clean build files"
	@echo "  deps                 - Install dependencies"
	@echo "  setup-config         - Create config.toml from template"
	@echo "  check-config         - Validate configuration"
	@echo "  clean-config         - Remove config.toml"
	@echo "  test                 - Run tests"
	@echo "  build-win            - Build for Windows"
	@echo "  build-linux          - Build for Linux"
	@echo "  build-mac            - Build for macOS"
	@echo "  install              - Install required packages"
	@echo ""
	@echo "Docker targets:"
	@echo "  docker-build         - Build Docker image"
	@echo "  docker-run           - Run Docker container"
	@echo "  docker-stop          - Stop and remove Docker container"
	@echo "  docker-logs          - Show Docker container logs"
	@echo "  docker-compose-up    - Start services with docker-compose"
	@echo "  docker-compose-down  - Stop services with docker-compose"
	@echo "  docker-compose-logs  - Show docker-compose logs"
	@echo "  setup-docker-env     - Setup Docker environment files"
	@echo "  docker-clean         - Clean Docker images and containers"
	@echo "  help                 - Show this help message"