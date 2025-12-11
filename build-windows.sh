#!/bin/bash
# Build script untuk Windows (cross-compile dari Linux)
# Requires: mingw-w64

set -e

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$PROJECT_DIR/backend"
ELECTRON_DIR="$PROJECT_DIR/electron-main"

print_info() { echo "ℹ $1"; }
print_error() { echo "✗ $1"; }
print_success() { echo "✓ $1"; }

print_info "Cross-compiling backend for Windows..."

# Check if mingw-w64 is installed
if ! command -v x86_64-w64-mingw32-gcc &> /dev/null; then
    print_error "mingw-w64 not found!"
    echo ""
    echo "Install with:"
    echo "  Ubuntu/Debian: sudo apt install gcc-mingw-w64-x86-64"
    echo "  Arch/Manjaro:  sudo pacman -S mingw-w64-gcc"
    echo "  Fedora:        sudo dnf install mingw64-gcc"
    exit 1
fi

cd "$BACKEND_DIR"

# Check Go installation
if ! command -v go &> /dev/null; then
    print_error "Go is not installed"
    exit 1
fi

# Download modules
print_info "Downloading Go modules..."
go mod download

# Build Windows binary with CGO enabled
print_info "Building Windows binary with CGO support..."
GOOS=windows \
GOARCH=amd64 \
CGO_ENABLED=1 \
CC=x86_64-w64-mingw32-gcc \
CXX=x86_64-w64-mingw32-g++ \
go build -o server.exe main.go

if [ $? -eq 0 ]; then
    print_success "Windows binary compiled: server.exe"
    
    # Copy to electron resources
    mkdir -p "$ELECTRON_DIR/resources/backend"
    cp "$BACKEND_DIR/server.exe" "$ELECTRON_DIR/resources/backend/"
    print_success "Windows binary copied to electron resources"
    
    print_info ""
    print_info "To build Windows installer, you need:"
    print_info "1. Wine installed on Linux, OR"
    print_info "2. Use a Windows machine, OR"
    print_info "3. Use GitHub Actions for CI/CD"
    
    # Option: Try to build with Wine if available
    if command -v wine &> /dev/null; then
        print_info ""
        print_info "Wine detected, attempting Windows installer build..."
        cd "$ELECTRON_DIR"
        npm ci
        npm run build:renderer && npx electron-builder --win nsis --publish never
        print_success "Windows installer created!"
    else
        print_info ""
        print_info "Wine not installed. Install with:"
        print_info "  Ubuntu/Debian: sudo apt install wine wine64"
        print_info "  Arch/Manjaro:  sudo pacman -S wine wine-mono wine-gecko"
    fi
else
    print_error "Failed to build Windows binary"
    exit 1
fi
