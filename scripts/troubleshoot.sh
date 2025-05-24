#!/bin/bash

# MyNodeCP Troubleshooting Script
# This script helps diagnose and fix common installation issues

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

# Check system status
check_system_status() {
    log_info "Checking system status..."
    
    echo "=== System Information ==="
    uname -a
    echo
    
    echo "=== Disk Space ==="
    df -h
    echo
    
    echo "=== Memory Usage ==="
    free -h
    echo
    
    echo "=== Load Average ==="
    uptime
    echo
}

# Check service status
check_services() {
    log_info "Checking service status..."
    
    echo "=== MyNodeCP Service ==="
    systemctl status mynodecp --no-pager || log_warning "MyNodeCP service not running"
    echo
    
    echo "=== MariaDB Service ==="
    systemctl status mariadb --no-pager || log_warning "MariaDB service not running"
    echo
    
    echo "=== Redis Service ==="
    if systemctl list-unit-files | grep -q "redis-server.service"; then
        systemctl status redis-server --no-pager || log_warning "Redis service not running"
    else
        systemctl status redis --no-pager || log_warning "Redis service not running"
    fi
    echo
    
    echo "=== Nginx Service ==="
    systemctl status nginx --no-pager || log_warning "Nginx service not running"
    echo
}

# Check network connectivity
check_network() {
    log_info "Checking network connectivity..."
    
    echo "=== Port Status ==="
    netstat -tlnp | grep -E ":80|:443|:8080|:9090|:3306|:6379" || log_warning "Some ports may not be listening"
    echo
    
    echo "=== Firewall Status ==="
    if command -v ufw &> /dev/null; then
        ufw status
    elif command -v firewall-cmd &> /dev/null; then
        firewall-cmd --list-all
    fi
    echo
}

# Check database connectivity
check_database() {
    log_info "Checking database connectivity..."
    
    if [[ -f "$MYNODECP_CONFIG/database.conf" ]]; then
        source "$MYNODECP_CONFIG/database.conf"
        
        echo "=== Database Connection Test ==="
        mysql -u "$DB_USER" -p"$DB_PASSWORD" -h "$DB_HOST" -e "SELECT 1;" 2>/dev/null && \
            log_success "Database connection successful" || \
            log_error "Database connection failed"
        
        echo "=== Database Tables ==="
        mysql -u "$DB_USER" -p"$DB_PASSWORD" -h "$DB_HOST" "$DB_NAME" -e "SHOW TABLES;" 2>/dev/null || \
            log_warning "Could not list database tables"
    else
        log_error "Database configuration file not found: $MYNODECP_CONFIG/database.conf"
    fi
    echo
}

# Check Redis connectivity
check_redis() {
    log_info "Checking Redis connectivity..."
    
    echo "=== Redis Connection Test ==="
    redis-cli ping 2>/dev/null && \
        log_success "Redis connection successful" || \
        log_error "Redis connection failed"
    echo
}

# Check file permissions
check_permissions() {
    log_info "Checking file permissions..."
    
    echo "=== MyNodeCP Directory Permissions ==="
    ls -la "$MYNODECP_HOME" 2>/dev/null || log_error "MyNodeCP directory not found"
    echo
    
    echo "=== Configuration Directory Permissions ==="
    ls -la "$MYNODECP_CONFIG" 2>/dev/null || log_error "Configuration directory not found"
    echo
    
    echo "=== Binary Permissions ==="
    ls -la "$MYNODECP_HOME/backend/mynodecp-server" 2>/dev/null || log_error "MyNodeCP binary not found"
    echo
}

# Check logs
check_logs() {
    log_info "Checking recent logs..."
    
    echo "=== MyNodeCP Service Logs (last 20 lines) ==="
    journalctl -u mynodecp -n 20 --no-pager || log_warning "Could not read MyNodeCP logs"
    echo
    
    echo "=== System Logs (errors in last hour) ==="
    journalctl --since "1 hour ago" -p err --no-pager | tail -10 || log_warning "Could not read system error logs"
    echo
}

# Fix common issues
fix_permissions() {
    log_info "Fixing file permissions..."
    
    if [[ -d "$MYNODECP_HOME" ]]; then
        chown -R $MYNODECP_USER:$MYNODECP_USER "$MYNODECP_HOME"
        chmod +x "$MYNODECP_HOME/backend/mynodecp-server" 2>/dev/null || true
        log_success "Fixed MyNodeCP directory permissions"
    fi
    
    if [[ -d "$MYNODECP_CONFIG" ]]; then
        chown -R $MYNODECP_USER:$MYNODECP_USER "$MYNODECP_CONFIG"
        chmod 600 "$MYNODECP_CONFIG"/*.conf 2>/dev/null || true
        log_success "Fixed configuration directory permissions"
    fi
}

# Restart services
restart_services() {
    log_info "Restarting services..."
    
    systemctl restart mariadb && log_success "MariaDB restarted" || log_error "Failed to restart MariaDB"
    
    if systemctl list-unit-files | grep -q "redis-server.service"; then
        systemctl restart redis-server && log_success "Redis restarted" || log_error "Failed to restart Redis"
    else
        systemctl restart redis && log_success "Redis restarted" || log_error "Failed to restart Redis"
    fi
    
    systemctl restart nginx && log_success "Nginx restarted" || log_error "Failed to restart Nginx"
    systemctl restart mynodecp && log_success "MyNodeCP restarted" || log_error "Failed to restart MyNodeCP"
}

# Show help
show_help() {
    echo "MyNodeCP Troubleshooting Script"
    echo
    echo "Usage: $0 [OPTION]"
    echo
    echo "Options:"
    echo "  check       Run all diagnostic checks"
    echo "  services    Check service status"
    echo "  network     Check network connectivity"
    echo "  database    Check database connectivity"
    echo "  redis       Check Redis connectivity"
    echo "  permissions Check file permissions"
    echo "  logs        Show recent logs"
    echo "  fix         Fix common permission issues"
    echo "  restart     Restart all services"
    echo "  help        Show this help message"
    echo
}

# Main function
main() {
    check_root
    
    case "${1:-check}" in
        check)
            check_system_status
            check_services
            check_network
            check_database
            check_redis
            check_permissions
            check_logs
            ;;
        services)
            check_services
            ;;
        network)
            check_network
            ;;
        database)
            check_database
            ;;
        redis)
            check_redis
            ;;
        permissions)
            check_permissions
            ;;
        logs)
            check_logs
            ;;
        fix)
            fix_permissions
            ;;
        restart)
            restart_services
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
