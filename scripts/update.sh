#!/bin/bash

# MyNodeCP Update Script
# This script updates MyNodeCP to the latest version

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
MYNODECP_USER="mynodecp"
MYNODECP_HOME="/opt/mynodecp"
MYNODECP_CONFIG="/etc/mynodecp"
BACKUP_DIR="/opt/mynodecp-backup-$(date +%Y%m%d-%H%M%S)"
GITHUB_REPO="mynodecp/mynodecp"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "This script must be run as root"
        exit 1
    fi
}

# Get current version
get_current_version() {
    if [[ -f "$MYNODECP_HOME/VERSION" ]]; then
        CURRENT_VERSION=$(cat "$MYNODECP_HOME/VERSION")
    else
        CURRENT_VERSION="unknown"
    fi
    log_info "Current version: $CURRENT_VERSION"
}

# Get latest version from GitHub
get_latest_version() {
    log_info "Checking for latest version..."
    LATEST_VERSION=$(curl -s "https://api.github.com/repos/$GITHUB_REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [[ -z "$LATEST_VERSION" ]]; then
        log_error "Could not fetch latest version from GitHub"
        exit 1
    fi
    
    log_info "Latest version: $LATEST_VERSION"
}

# Check if update is needed
check_update_needed() {
    if [[ "$CURRENT_VERSION" == "$LATEST_VERSION" ]]; then
        log_success "MyNodeCP is already up to date!"
        exit 0
    fi
    
    log_info "Update available: $CURRENT_VERSION -> $LATEST_VERSION"
}

# Create backup
create_backup() {
    log_info "Creating backup..."
    
    mkdir -p "$BACKUP_DIR"
    
    # Backup application files
    cp -r "$MYNODECP_HOME" "$BACKUP_DIR/mynodecp"
    
    # Backup configuration
    cp -r "$MYNODECP_CONFIG" "$BACKUP_DIR/config"
    
    # Backup database
    if [[ -f "$MYNODECP_CONFIG/database.conf" ]]; then
        source "$MYNODECP_CONFIG/database.conf"
        mysqldump -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" > "$BACKUP_DIR/database.sql"
    fi
    
    log_success "Backup created at: $BACKUP_DIR"
}

# Stop services
stop_services() {
    log_info "Stopping MyNodeCP services..."
    
    systemctl stop mynodecp || log_warning "Could not stop MyNodeCP service"
    systemctl stop nginx || log_warning "Could not stop Nginx service"
    
    log_success "Services stopped"
}

# Download and extract update
download_update() {
    log_info "Downloading MyNodeCP $LATEST_VERSION..."
    
    cd /tmp
    wget "https://github.com/$GITHUB_REPO/archive/refs/tags/$LATEST_VERSION.tar.gz" -O "mynodecp-$LATEST_VERSION.tar.gz"
    
    tar -xzf "mynodecp-$LATEST_VERSION.tar.gz"
    
    log_success "Update downloaded and extracted"
}

# Install update
install_update() {
    log_info "Installing update..."
    
    # Remove old application files (keep config and data)
    rm -rf "$MYNODECP_HOME/backend" "$MYNODECP_HOME/frontend" "$MYNODECP_HOME/scripts"
    
    # Copy new files
    cp -r "/tmp/mynodecp-$LATEST_VERSION"/* "$MYNODECP_HOME/"
    
    # Build backend
    cd "$MYNODECP_HOME/backend"
    export PATH=$PATH:/usr/local/go/bin
    go mod download
    go build -o mynodecp-server cmd/server/main.go
    
    # Build frontend
    cd "$MYNODECP_HOME/frontend"
    npm install
    npm run build
    
    # Set permissions
    chown -R $MYNODECP_USER:$MYNODECP_USER "$MYNODECP_HOME"
    chmod +x "$MYNODECP_HOME/backend/mynodecp-server"
    
    # Update version file
    echo "$LATEST_VERSION" > "$MYNODECP_HOME/VERSION"
    
    log_success "Update installed"
}

# Run database migrations
run_migrations() {
    log_info "Running database migrations..."
    
    cd "$MYNODECP_HOME/backend"
    sudo -u $MYNODECP_USER ./mynodecp-server --migrate || log_warning "Migration command not available"
    
    log_success "Database migrations completed"
}

# Start services
start_services() {
    log_info "Starting MyNodeCP services..."
    
    systemctl daemon-reload
    systemctl start mynodecp || log_error "Could not start MyNodeCP service"
    systemctl start nginx || log_error "Could not start Nginx service"
    
    # Wait for services to start
    sleep 5
    
    # Check if services are running
    if systemctl is-active --quiet mynodecp; then
        log_success "MyNodeCP service started successfully"
    else
        log_error "MyNodeCP service failed to start"
        log_info "Check logs with: journalctl -u mynodecp -f"
        exit 1
    fi
    
    log_success "Services started"
}

# Verify update
verify_update() {
    log_info "Verifying update..."
    
    # Check if MyNodeCP is responding
    sleep 10
    if curl -f http://localhost:8080/health >/dev/null 2>&1; then
        log_success "MyNodeCP is responding correctly"
    else
        log_warning "MyNodeCP may not be responding correctly"
        log_info "Check the service status and logs"
    fi
    
    log_success "Update verification completed"
}

# Cleanup
cleanup() {
    log_info "Cleaning up temporary files..."
    
    rm -rf "/tmp/mynodecp-$LATEST_VERSION"
    rm -f "/tmp/mynodecp-$LATEST_VERSION.tar.gz"
    
    log_success "Cleanup completed"
}

# Rollback function
rollback() {
    log_warning "Rolling back to previous version..."
    
    if [[ ! -d "$BACKUP_DIR" ]]; then
        log_error "Backup directory not found: $BACKUP_DIR"
        exit 1
    fi
    
    # Stop services
    systemctl stop mynodecp || true
    
    # Restore files
    rm -rf "$MYNODECP_HOME"
    cp -r "$BACKUP_DIR/mynodecp" "$MYNODECP_HOME"
    
    # Restore database
    if [[ -f "$BACKUP_DIR/database.sql" && -f "$MYNODECP_CONFIG/database.conf" ]]; then
        source "$MYNODECP_CONFIG/database.conf"
        mysql -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < "$BACKUP_DIR/database.sql"
    fi
    
    # Start services
    systemctl start mynodecp
    
    log_success "Rollback completed"
}

# Show help
show_help() {
    echo "MyNodeCP Update Script"
    echo
    echo "Usage: $0 [OPTION]"
    echo
    echo "Options:"
    echo "  update      Update MyNodeCP to the latest version (default)"
    echo "  check       Check for available updates"
    echo "  rollback    Rollback to the previous version"
    echo "  help        Show this help message"
    echo
}

# Main function
main() {
    check_root
    
    case "${1:-update}" in
        update)
            get_current_version
            get_latest_version
            check_update_needed
            create_backup
            stop_services
            download_update
            install_update
            run_migrations
            start_services
            verify_update
            cleanup
            
            echo
            log_success "MyNodeCP has been successfully updated to version $LATEST_VERSION!"
            log_info "Backup created at: $BACKUP_DIR"
            log_info "If you encounter any issues, you can rollback using: $0 rollback"
            ;;
        check)
            get_current_version
            get_latest_version
            check_update_needed
            ;;
        rollback)
            rollback
            ;;
        help)
            show_help
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
