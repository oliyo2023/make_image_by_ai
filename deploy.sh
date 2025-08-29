#!/bin/bash

# AI Image Generator Docker deployment script
# Usage: ./deploy.sh [command]
# Commands: setup, build, run, stop, logs, clean

set -e

# Project configuration
PROJECT_NAME="ai-image-generator"
IMAGE_NAME="ai-image-generator"
CONTAINER_NAME="ai-image-generator"
PORT="8000"

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored messages
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is installed
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed, please install Docker first"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        print_error "Docker service is not running, please start Docker service"
        exit 1
    fi
    
    print_success "Docker check passed"
}

# Setup environment
setup_environment() {
    print_info "Setting up Docker environment..."
    
    # Create environment variable file
    if [ ! -f ".env" ]; then
        cp .env.example .env
        print_success "Created .env file, please edit this file to set your API keys"
        print_warning "Please set the following required environment variables in the .env file:"
        echo "  - MODEL_SCOPE_TOKEN"
        echo "  - OPENROUTER_API_KEY"
    else
        print_info ".env file already exists"
    fi
    
    # Create necessary directories
    mkdir -p public/static/images
    mkdir -p logs
    print_success "Created necessary directories"
    
    print_success "Environment setup completed"
}

# Build image
build_image() {
    print_info "Building Docker image..."
    
    docker build -t ${IMAGE_NAME}:latest .
    
    print_success "Image build completed"
}

# Run container
run_container() {
    print_info "Starting container..."
    
    # Stop existing container
    stop_container
    
    # Start new container
    docker run -d \
        --name ${CONTAINER_NAME} \
        -p ${PORT}:${PORT} \
        -v "$(pwd)/public/static/images:/app/public/static/images" \
        -v "$(pwd)/logs:/app/logs" \
        --env-file .env \
        --restart unless-stopped \
        ${IMAGE_NAME}:latest
    
    print_success "Container started successfully"
    print_info "Service URL: http://localhost:${PORT}"
    print_info "Health check: http://localhost:${PORT}/health"
    
    # Wait for service to start
    print_info "Waiting for service to start..."
    sleep 5
    
    # Check container status
    if docker ps | grep -q ${CONTAINER_NAME}; then
        print_success "Container is running normally"
    else
        print_error "Container failed to start, check logs:"
        docker logs ${CONTAINER_NAME}
        exit 1
    fi
}

# Stop container
stop_container() {
    if docker ps -a | grep -q ${CONTAINER_NAME}; then
        print_info "Stopping container..."
        docker stop ${CONTAINER_NAME} || true
        docker rm ${CONTAINER_NAME} || true
        print_success "Container stopped"
    fi
}

# Show logs
show_logs() {
    if docker ps | grep -q ${CONTAINER_NAME}; then
        print_info "Showing container logs (Press Ctrl+C to exit)..."
        docker logs -f ${CONTAINER_NAME}
    else
        print_error "Container is not running"
        exit 1
    fi
}

# Clean resources
clean_resources() {
    print_info "Cleaning Docker resources..."
    
    # Stop container
    stop_container
    
    # Remove image
    if docker images | grep -q ${IMAGE_NAME}; then
        docker rmi ${IMAGE_NAME}:latest
        print_success "Image removed"
    fi
    
    # Clean system
    docker system prune -f
    print_success "Docker resource cleanup completed"
}

# Show status
show_status() {
    print_info "Docker Status:"
    
    echo "=== Images ==="
    docker images | grep ${IMAGE_NAME} || echo "No related images"
    
    echo "=== Containers ==="
    docker ps -a | grep ${CONTAINER_NAME} || echo "No related containers"
    
    echo "=== Running Status ==="
    if docker ps | grep -q ${CONTAINER_NAME}; then
        print_success "Container is running"
        echo "Service URL: http://localhost:${PORT}"
    else
        print_warning "Container is not running"
    fi
}

# Usage help
show_help() {
    echo "AI Image Generator Docker Deployment Script"
    echo ""
    echo "Usage:"
    echo "  $0 [command]"
    echo ""
    echo "Available commands:"
    echo "  setup   - Setup environment and config files"
    echo "  build   - Build Docker image"
    echo "  run     - Run container"
    echo "  stop    - Stop container"
    echo "  logs    - Show container logs"
    echo "  status  - Show status"
    echo "  clean   - Clean all resources"
    echo "  help    - Show this help message"
    echo ""
    echo "Quick start:"
    echo "  1. $0 setup    # Setup environment"
    echo "  2. Edit .env file, set API keys"
    echo "  3. $0 build    # Build image"
    echo "  4. $0 run      # Run service"
}

# Main function
main() {
    case "${1:-help}" in
        setup)
            check_docker
            setup_environment
            ;;
        build)
            check_docker
            build_image
            ;;
        run)
            check_docker
            run_container
            ;;
        stop)
            check_docker
            stop_container
            ;;
        logs)
            check_docker
            show_logs
            ;;
        status)
            check_docker
            show_status
            ;;
        clean)
            check_docker
            clean_resources
            ;;
        help|*)
            show_help
            ;;
    esac
}

# Execute main function
main "$@"