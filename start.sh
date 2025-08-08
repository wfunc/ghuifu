#!/bin/bash

# Huifu Configuration System Startup Script

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_message() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        print_message $RED "Error: Go is not installed!"
        print_message $YELLOW "Please install Go from https://golang.org/dl/"
        exit 1
    fi
    print_message $GREEN "✓ Go is installed: $(go version)"
}

# Install dependencies
install_deps() {
    print_message $YELLOW "Installing dependencies..."
    go mod download
    go mod tidy
    print_message $GREEN "✓ Dependencies installed"
}

# Build the application
build_app() {
    print_message $YELLOW "Building application..."
    go build -o huifu-server main.go
    if [ $? -eq 0 ]; then
        print_message $GREEN "✓ Build successful"
    else
        print_message $RED "✗ Build failed"
        exit 1
    fi
}

# Start the server
start_server() {
    local mode=$1
    
    if [ "$mode" == "dev" ]; then
        print_message $YELLOW "Starting server in development mode..."
        print_message $GREEN "Server URL: http://localhost:8080"
        print_message $YELLOW "Press Ctrl+C to stop"
        go run main.go
    elif [ "$mode" == "prod" ]; then
        print_message $YELLOW "Starting server in production mode..."
        
        # Check if binary exists
        if [ ! -f "./huifu-server" ]; then
            build_app
        fi
        
        print_message $GREEN "Server URL: http://localhost:8080"
        print_message $YELLOW "Press Ctrl+C to stop"
        ./huifu-server
    elif [ "$mode" == "docker" ]; then
        print_message $YELLOW "Starting with Docker Compose..."
        
        # Check if docker-compose is installed
        if ! command -v docker-compose &> /dev/null; then
            print_message $RED "Error: Docker Compose is not installed!"
            exit 1
        fi
        
        docker-compose up -d
        print_message $GREEN "✓ Docker containers started"
        print_message $GREEN "Server URL: http://localhost:8080"
        print_message $YELLOW "Run 'docker-compose logs -f' to view logs"
        print_message $YELLOW "Run 'docker-compose down' to stop"
    else
        print_message $RED "Invalid mode: $mode"
        print_message $YELLOW "Usage: ./start.sh [dev|prod|docker]"
        exit 1
    fi
}

# Main script
main() {
    print_message $GREEN "==================================="
    print_message $GREEN "  Huifu Configuration System"
    print_message $GREEN "==================================="
    
    # Default mode is development
    MODE=${1:-dev}
    
    if [ "$MODE" != "docker" ]; then
        check_go
        install_deps
    fi
    
    start_server $MODE
}

# Handle script arguments
case "$1" in
    dev|prod|docker)
        main $1
        ;;
    help|--help|-h)
        echo "Huifu Configuration System Startup Script"
        echo ""
        echo "Usage: ./start.sh [mode]"
        echo ""
        echo "Modes:"
        echo "  dev     Start in development mode (default)"
        echo "  prod    Build and start in production mode"
        echo "  docker  Start using Docker Compose"
        echo ""
        echo "Examples:"
        echo "  ./start.sh         # Start in development mode"
        echo "  ./start.sh dev     # Start in development mode"
        echo "  ./start.sh prod    # Start in production mode"
        echo "  ./start.sh docker  # Start with Docker"
        ;;
    *)
        main "dev"
        ;;
esac