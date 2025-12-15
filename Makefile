.PHONY: help install build-fe build-be build run-fe run-be run dev clean

help: ## Show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

install: ## Install dependencies for frontend and backend
	@echo "Installing frontend dependencies..."
	cd renderer && npm install
	@echo "Installing electron-main dependencies..."
	cd electron-main && npm install
	@echo "Installing backend dependencies..."
	cd backend && go mod download
	@echo "All dependencies installed."

build-fe: ## Build frontend (renderer) for production
	@echo "Building frontend..."
	cd renderer && npm run build
	@echo "Frontend build complete."

build-be: ## Build backend Go binary (server)
	@echo "Building backend..."
	cd backend && go build -o server main.go
	@echo "Backend build complete: backend/server"

build: build-fe build-be ## Build both frontend and backend

run-fe: ## Run frontend dev server (Vite on :5173)
	@echo "Starting frontend dev server..."
	cd renderer && npm run dev

run-be: ## Run backend server (Go on :8080)
	@echo "Starting backend server..."
	cd backend && go run main.go

run: ## Run both frontend and backend in development mode (requires tmux or separate terminals)
	@echo "To run both, open two terminals:"
	@echo "  Terminal 1: make run-be"
	@echo "  Terminal 2: make run-fe"
	@echo "Or use tmux/screen to run concurrently."

dev: ## Run backend and frontend concurrently (requires 'concurrently' or manual setup)
	@echo "Run 'make run-be' and 'make run-fe' in separate terminals."

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf renderer/dist
	rm -f backend/server backend/server.exe
	rm -rf release/linux-unpacked release/win-unpacked release/*.AppImage release/*.exe release/*.blockmap
	@echo "Clean complete."

dist-linux: ## Build Electron app for Linux
	@echo "Building Electron app for Linux..."
	cd electron-main && npm run dist:linux
	@echo "Linux build complete in release/"

dist-win: ## Build Electron app for Windows
	@echo "Building Electron app for Windows..."
	cd electron-main && npm run dist:win
	@echo "Windows build complete in release/"

dist: ## Build Electron app for all platforms
	@echo "Building Electron app for all platforms..."
	cd electron-main && npm run dist
	@echo "Distribution build complete in release/"
