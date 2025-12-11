#!/bin/bash
# Build script untuk Shosha Mart Electron App
# Supports: Linux, Windows, macOS

set -e

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$PROJECT_DIR/backend"
RENDERER_DIR="$PROJECT_DIR/renderer"
ELECTRON_DIR="$PROJECT_DIR/electron-main"
RELEASE_DIR="$PROJECT_DIR/release"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_info() { echo -e "${GREEN}ℹ${NC} $1"; }
print_warn() { echo -e "${YELLOW}⚠${NC} $1"; }
print_error() { echo -e "${RED}✗${NC} $1"; }
print_success() { echo -e "${GREEN}✓${NC} $1"; }

# Check OS
OS=$(uname -s)
ARCH=$(uname -m)

print_info "Building for OS: $OS ($ARCH)"
print_info "Build directory: $RELEASE_DIR"

# 1. Build backend binary
print_info "Building backend binary..."
cd "$BACKEND_DIR"

if ! command -v go &> /dev/null; then
    print_error "Go is not installed"
    exit 1
fi

go mod download
CGO_ENABLED=1 go build -o server main.go
print_success "Backend binary compiled: $BACKEND_DIR/server"

# 2. Build renderer
print_info "Building renderer (Vue)..."
cd "$RENDERER_DIR"

if ! command -v npm &> /dev/null; then
    print_error "npm is not installed"
    exit 1
fi

npm ci
npm run build
print_success "Renderer built: $RENDERER_DIR/dist"

# 3. Prepare backend for electron packaging
print_info "Preparing backend for packaging..."
mkdir -p "$ELECTRON_DIR/resources/backend"
cp "$BACKEND_DIR/server" "$ELECTRON_DIR/resources/backend/" || \
cp "$BACKEND_DIR/server.exe" "$ELECTRON_DIR/resources/backend/" 2>/dev/null || true
print_success "Backend binary copied to electron resources"

# 4. Build electron app
print_info "Building Electron app..."
cd "$ELECTRON_DIR"

npm ci

case "$OS" in
    Linux)
        print_info "Building for Linux (AppImage)..."
        npm run dist:linux
        print_success "Linux AppImage created"
        ;;
    Darwin)
        print_info "Building for macOS (DMG + ZIP)..."
        npm run build:renderer && npx electron-builder --mac --publish never
        print_success "macOS build created"
        ;;
    MINGW*|MSYS*|CYGWIN*)
        print_info "Building for Windows (NSIS + Portable)..."
        npm run dist:win
        print_success "Windows installer created"
        ;;
    *)
        print_error "Unsupported OS: $OS"
        exit 1
        ;;
esac

print_success "Build completed successfully!"
print_info "Outputs in: $RELEASE_DIR"
ls -lh "$RELEASE_DIR" | grep -E '\.(AppImage|exe|dmg|zip)$' || true
