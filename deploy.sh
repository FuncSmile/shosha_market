#!/bin/bash

# ========================================
# ShoshaMart Deployment Script
# ========================================
# This script helps deploy ShoshaMart to production server
# Usage: ./deploy.sh [start|stop|restart|logs|status]

set -e

# Colors untuk output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker-compose.prod.yml"
ENV_FILE=".env.production.local"  # Use .local for production secrets
PROJECT_NAME="shosha-mart"

# Functions
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

check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed!"
        exit 1
    fi
    
    if ! docker compose version &> /dev/null; then
        print_error "Docker Compose is not installed!"
        exit 1
    fi
    
    print_success "Docker and Docker Compose are installed"
}

check_env_file() {
    if [ ! -f "$ENV_FILE" ]; then
        print_error "Environment file $ENV_FILE not found!"
        print_info "Creating from template..."
        
        if [ ! -f ".env.production" ]; then
            print_error ".env.production template not found!"
            exit 1
        fi
        
        cp .env.production "$ENV_FILE"
        chmod 600 "$ENV_FILE"
        
        print_warning "Created $ENV_FILE with template values"
        print_error "⚠️  STOP! Update secrets in $ENV_FILE before deploying!"
        print_info "Required changes:"
        echo "  1. POSTGRES_PASSWORD"
        echo "  2. JWT_SECRET (run: openssl rand -base64 64)"
        echo "  3. MINIO_ROOT_PASSWORD"
    # Create required directories for volume mounts
    print_info "Creating required directories..."
    mkdir -p data/postgres data/minio data/uploads logs/nginx
    
    # Set proper permissions
    chmod 700 data/postgres data/minio
    chmod 755 data/uploads logs/nginx
    
        echo "  4. Update YOUR_SERVER_IP and YOUR_DOMAIN_OR_IP"
        exit 1
    fi
    
    # Check file permissions
    PERMS=$(stat -c "%a" "$ENV_FILE" 2>/dev/null || stat -f "%Lp" "$ENV_FILE" 2>/dev/null)
    if [ "$PERMS" != "600" ]; then
        print_warning "Fixing $ENV_FILE permissions (should be 600)..."
        chmod 600 "$ENV_FILE"
    fi
    
    # Check untuk placeholder values yang harus diganti
    if grep -q "YOUR_SERVER_IP\|YOUR_DOMAIN_OR_IP\|YOUR_STRONG_PASSWORD_HERE_CHANGE_ME\|YOUR_SUPER_SECRET_JWT_KEY_CHANGE_THIS" "$ENV_FILE"; then
        print_error "⚠️  Found placeholder values in $ENV_FILE!"
        print_error "Please update these before deploying:"
        grep -n "YOUR_" "$ENV_FILE" | head -5
        exit 1
    fi
    
    # Validate required variables are set
    print_info "Validating environment variables..."
    
    REQUIRED_VARS="POSTGRES_PASSWORD JWT_SECRET MINIO_ROOT_PASSWORD FRONTEND_URL NUXT_PUBLIC_API_URL"
    for var in $REQUIRED_VARS; do
        if ! grep -q "^${var}=.\+" "$ENV_FILE"; then
            print_error "Required variable $var not set in $ENV_FILE"
            exit 1
        fi
    done
    
    print_success "Environment file validated"
}

start_services() {
    print_info "Starting ShoshaMart services..."
    
    docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" up -d
    
    print_success "Services started successfully!"
    print_info "Waiting for services to be ready..."
    sleep 10
    
    show_status
}

stop_services() {
    print_info "Stopping ShoshaMart services..."
    
    docker compose -f "$COMPOSE_FILE" down
    
    print_success "Services stopped successfully!"
}

restart_services() {
    print_info "Restarting ShoshaMart services..."
    
    stop_services
    sleep 3
    start_services
}

show_logs() {
    print_info "Showing logs (Ctrl+C to exit)..."
    docker compose -f "$COMPOSE_FILE" logs -f --tail=100
}

show_status() {
    print_info "Service Status:"
    echo ""
    docker compose -f "$COMPOSE_FILE" ps
    echo ""
    
    # Check health status from Docker
    print_info "Health Check (Docker Status):"
    
    services="backend frontend postgres minio nginx"
    
    for service in $services; do
        container_name="${PROJECT_NAME}_${service//-/_}"
        # Handle special naming if needed, but project name is shosha-mart -> shosha_mart
        # Service names in compose are simple.
        # Container names are defined in compose file: shosha_mart_backend, etc.
        
        status=$(docker inspect --format='{{.State.Health.Status}}' "shosha_mart_$service" 2>/dev/null)
        
        if [ "$status" == "healthy" ]; then
            print_success "✓ $service is healthy"
        elif [ "$status" == "unhealthy" ]; then
            print_error "✗ $service is unhealthy"
        elif [ "$status" == "starting" ]; then
            print_warning "⧖ $service is starting"
        else
            print_warning "? $service status: ${status:-unknown}"
        fi
    done
}

build_images() {
    print_info "Building Docker images..."
    
    docker compose -f "$COMPOSE_FILE" --env-file "$ENV_FILE" build --no-cache
    
    print_success "Images built successfully!"
}

clean_volumes() {
    print_warning "This will delete all data (database, uploads, minio files)!"
    read -p "Are you sure? (yes/no): " confirm
    
    if [ "$confirm" = "yes" ]; then
        print_info "Cleaning volumes..."
        docker compose -f "$COMPOSE_FILE" down -v
        print_success "Volumes cleaned!"
    else
        print_info "Cancelled"
    fi
}

backup_data() {
    print_info "Creating backup..."
    
    BACKUP_DIR="backups/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$BACKUP_DIR"
    
    # Backup PostgreSQL
    print_info "Backing up PostgreSQL..."
    docker compose -f "$COMPOSE_FILE" exec -T postgres pg_dump -U postgres shosha_mart > "$BACKUP_DIR/postgres.sql"
    
    # Backup MinIO data
    print_info "Backing up MinIO..."
    docker cp shosha_mart_minio:/data "$BACKUP_DIR/minio-data"
    
    print_success "Backup created at $BACKUP_DIR"
}

show_help() {
    cat << EOF
${GREEN}ShoshaMart Deployment Script${NC}

Usage: ./deploy.sh [command]

Commands:
    start       Start all services
    stop        Stop all services
    restart     Restart all services
    logs        Show logs (tail -f)
    status      Show service status
    build       Build Docker images
    clean       Clean volumes (WARNING: deletes all data!)
    backup      Backup database and files
    help        Show this help message

Examples:
    ./deploy.sh start       # Start services
    ./deploy.sh logs        # View logs
    ./deploy.sh status      # Check status

EOF
}

# Main script
main() {
    print_info "ShoshaMart Deployment Manager"
    echo ""
    
    # Check prerequisites
    check_docker
    
    # Parse command
    case "${1:-help}" in
        start)
            check_env_file
            start_services
            ;;
        stop)
            stop_services
            ;;
        restart)
            check_env_file
            restart_services
            ;;
        logs)
            show_logs
            ;;
        status)
            show_status
            ;;
        build)
            check_env_file
            build_images
            ;;
        clean)
            clean_volumes
            ;;
        backup)
            backup_data
            ;;
        help|*)
            show_help
            ;;
    esac
}

# Run main
main "$@"
