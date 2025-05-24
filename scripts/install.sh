#!/bin/bash

# MyNodeCP Installation Script
# Enterprise-grade web hosting control panel installer
# Supports: RHEL, CentOS, Rocky Linux, AlmaLinux, Ubuntu LTS, Debian

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
MYNODECP_VERSION="1.0.0"
MYNODECP_USER="mynodecp"
MYNODECP_GROUP="mynodecp"
MYNODECP_HOME="/opt/mynodecp"
MYNODECP_CONFIG="/etc/mynodecp"
MYNODECP_LOGS="/var/log/mynodecp"
MYNODECP_DATA="/var/lib/mynodecp"

# Database configuration
DB_NAME="mynodecp"
DB_USER="mynodecp"
DB_PASSWORD=""
MYSQL_ROOT_PASSWORD="$(openssl rand -base64 32)"

# Service ports
HTTP_PORT=8080
GRPC_PORT=9090

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

# Detect operating system
detect_os() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        OS=$ID
        OS_VERSION=$VERSION_ID
    else
        log_error "Cannot detect operating system"
        exit 1
    fi

    log_info "Detected OS: $OS $OS_VERSION"

    case $OS in
        rhel|centos|rocky|almalinux)
            PACKAGE_MANAGER="yum"
            if command -v dnf &> /dev/null; then
                PACKAGE_MANAGER="dnf"
            fi
            ;;
        ubuntu|debian)
            PACKAGE_MANAGER="apt"
            ;;
        *)
            log_error "Unsupported operating system: $OS"
            exit 1
            ;;
    esac
}

# Update system packages
update_system() {
    log_info "Updating system packages..."

    case $PACKAGE_MANAGER in
        yum|dnf)
            $PACKAGE_MANAGER update -y
            $PACKAGE_MANAGER install -y epel-release
            ;;
        apt)
            apt update
            apt upgrade -y
            ;;
    esac

    log_success "System packages updated"
}

# Install dependencies
install_dependencies() {
    log_info "Installing system dependencies..."

    local packages=""

    case $PACKAGE_MANAGER in
        yum|dnf)
            packages="curl wget git unzip tar gzip openssl ca-certificates systemd firewalld"
            $PACKAGE_MANAGER install -y $packages
            ;;
        apt)
            packages="curl wget git unzip tar gzip openssl ca-certificates systemd ufw"
            apt install -y $packages
            ;;
    esac

    log_success "System dependencies installed"
}

# Install Go
install_go() {
    log_info "Installing Go 1.21..."

    local go_version="1.21.5"
    local go_archive="go${go_version}.linux-amd64.tar.gz"

    # Remove existing Go installation
    rm -rf /usr/local/go

    # Download and install Go
    cd /tmp
    wget "https://golang.org/dl/${go_archive}"
    tar -C /usr/local -xzf "${go_archive}"

    # Add Go to PATH
    echo 'export PATH=$PATH:/usr/local/go/bin' > /etc/profile.d/go.sh
    chmod +x /etc/profile.d/go.sh

    # Verify installation
    /usr/local/go/bin/go version

    log_success "Go installed successfully"
}

# Install Node.js
install_nodejs() {
    log_info "Installing Node.js 18..."

    # Install Node.js using NodeSource repository
    curl -fsSL https://deb.nodesource.com/setup_18.x | bash -

    case $PACKAGE_MANAGER in
        yum|dnf)
            curl -fsSL https://rpm.nodesource.com/setup_18.x | bash -
            $PACKAGE_MANAGER install -y nodejs
            ;;
        apt)
            apt install -y nodejs
            ;;
    esac

    # Verify installation
    node --version
    npm --version

    log_success "Node.js installed successfully"
}

# Install MariaDB
install_mariadb() {
    log_info "Installing MariaDB 10.11..."

    # Set non-interactive mode for MariaDB
    export DEBIAN_FRONTEND=noninteractive

    case $PACKAGE_MANAGER in
        yum|dnf)
            $PACKAGE_MANAGER install -y mariadb-server mariadb
            ;;
        apt)
            # Pre-configure MariaDB to avoid interactive prompts
            echo "mariadb-server mysql-server/root_password password $MYSQL_ROOT_PASSWORD" | debconf-set-selections
            echo "mariadb-server mysql-server/root_password_again password $MYSQL_ROOT_PASSWORD" | debconf-set-selections
            apt install -y mariadb-server mariadb-client
            ;;
    esac

    # Start and enable MariaDB
    systemctl start mariadb
    systemctl enable mariadb

    # Secure MariaDB installation non-interactively
    mysql -e "UPDATE mysql.user SET Password = PASSWORD('$MYSQL_ROOT_PASSWORD') WHERE User = 'root';" 2>/dev/null || \
    mysql -e "ALTER USER 'root'@'localhost' IDENTIFIED BY '$MYSQL_ROOT_PASSWORD';" 2>/dev/null || \
    mysqladmin -u root password "$MYSQL_ROOT_PASSWORD" 2>/dev/null

    mysql -u root -p"$MYSQL_ROOT_PASSWORD" -e "DELETE FROM mysql.user WHERE User='';" 2>/dev/null || true
    mysql -u root -p"$MYSQL_ROOT_PASSWORD" -e "DELETE FROM mysql.user WHERE User='root' AND Host NOT IN ('localhost', '127.0.0.1', '::1');" 2>/dev/null || true
    mysql -u root -p"$MYSQL_ROOT_PASSWORD" -e "DROP DATABASE IF EXISTS test;" 2>/dev/null || true
    mysql -u root -p"$MYSQL_ROOT_PASSWORD" -e "DELETE FROM mysql.db WHERE Db='test' OR Db='test\\_%';" 2>/dev/null || true
    mysql -u root -p"$MYSQL_ROOT_PASSWORD" -e "FLUSH PRIVILEGES;" 2>/dev/null || true

    log_success "MariaDB installed and secured"
}

# Install Redis
install_redis() {
    log_info "Installing Redis..."

    case $PACKAGE_MANAGER in
        yum|dnf)
            $PACKAGE_MANAGER install -y redis
            REDIS_SERVICE="redis"
            ;;
        apt)
            apt install -y redis-server
            REDIS_SERVICE="redis-server"
            ;;
    esac

    # Detect actual Redis service name
    if systemctl list-unit-files | grep -q "redis-server.service"; then
        REDIS_SERVICE="redis-server"
    elif systemctl list-unit-files | grep -q "redis.service"; then
        REDIS_SERVICE="redis"
    else
        log_warning "Could not detect Redis service name, trying both..."
        REDIS_SERVICE="redis"
    fi

    # Start and enable Redis
    systemctl start $REDIS_SERVICE || systemctl start redis-server || log_warning "Failed to start Redis service"
    systemctl enable $REDIS_SERVICE || systemctl enable redis-server || log_warning "Failed to enable Redis service"

    log_success "Redis installed and configured with service: $REDIS_SERVICE"
}

# Install Nginx
install_nginx() {
    log_info "Installing Nginx..."

    case $PACKAGE_MANAGER in
        yum|dnf)
            $PACKAGE_MANAGER install -y nginx
            ;;
        apt)
            apt install -y nginx
            ;;
    esac

    # Start and enable Nginx
    systemctl start nginx
    systemctl enable nginx

    log_success "Nginx installed and configured"
}

# Create MyNodeCP user and directories
create_user_and_directories() {
    log_info "Creating MyNodeCP user and directories..."

    # Create user and group
    if ! getent group $MYNODECP_GROUP > /dev/null; then
        groupadd $MYNODECP_GROUP
    fi

    if ! getent passwd $MYNODECP_USER > /dev/null; then
        useradd -r -g $MYNODECP_GROUP -d $MYNODECP_HOME -s /bin/bash $MYNODECP_USER
    fi

    # Create directories
    mkdir -p $MYNODECP_HOME
    mkdir -p $MYNODECP_CONFIG
    mkdir -p $MYNODECP_LOGS
    mkdir -p $MYNODECP_DATA
    mkdir -p /var/www

    # Set permissions
    chown -R $MYNODECP_USER:$MYNODECP_GROUP $MYNODECP_HOME
    chown -R $MYNODECP_USER:$MYNODECP_GROUP $MYNODECP_CONFIG
    chown -R $MYNODECP_USER:$MYNODECP_GROUP $MYNODECP_LOGS
    chown -R $MYNODECP_USER:$MYNODECP_GROUP $MYNODECP_DATA

    log_success "User and directories created"
}

# Setup database
setup_database() {
    log_info "Setting up MyNodeCP database..."

    # Generate random password if not set
    if [[ -z "$DB_PASSWORD" ]]; then
        DB_PASSWORD=$(openssl rand -base64 32)
    fi

    # Create database and user
    mysql -u root -p"$MYSQL_ROOT_PASSWORD" -e "CREATE DATABASE IF NOT EXISTS $DB_NAME;"
    mysql -u root -p"$MYSQL_ROOT_PASSWORD" -e "CREATE USER IF NOT EXISTS '$DB_USER'@'localhost' IDENTIFIED BY '$DB_PASSWORD';"
    mysql -u root -p"$MYSQL_ROOT_PASSWORD" -e "GRANT ALL PRIVILEGES ON $DB_NAME.* TO '$DB_USER'@'localhost';"
    mysql -u root -p"$MYSQL_ROOT_PASSWORD" -e "FLUSH PRIVILEGES;"

    # Save database credentials
    cat > $MYNODECP_CONFIG/database.conf << EOF
DB_HOST=localhost
DB_PORT=3306
DB_NAME=$DB_NAME
DB_USER=$DB_USER
DB_PASSWORD=$DB_PASSWORD
EOF

    chmod 600 $MYNODECP_CONFIG/database.conf
    chown $MYNODECP_USER:$MYNODECP_GROUP $MYNODECP_CONFIG/database.conf

    log_success "Database configured"
}

# Configure firewall
configure_firewall() {
    log_info "Configuring firewall..."

    case $OS in
        rhel|centos|rocky|almalinux)
            # Configure firewalld
            systemctl start firewalld
            systemctl enable firewalld

            firewall-cmd --permanent --add-port=80/tcp
            firewall-cmd --permanent --add-port=443/tcp
            firewall-cmd --permanent --add-port=$HTTP_PORT/tcp
            firewall-cmd --permanent --add-service=ssh
            firewall-cmd --reload
            ;;
        ubuntu|debian)
            # Configure ufw
            ufw --force enable
            ufw allow ssh
            ufw allow 80/tcp
            ufw allow 443/tcp
            ufw allow $HTTP_PORT/tcp
            ;;
    esac

    log_success "Firewall configured"
}

# Install MyNodeCP
install_mynodecp() {
    log_info "Installing MyNodeCP..."

    # Get current directory (where the script is running from)
    SCRIPT_PATH="${BASH_SOURCE[0]}"
    SCRIPT_DIR="$(dirname "$SCRIPT_PATH")"

    # Handle relative paths properly
    if [[ "$SCRIPT_DIR" == "." ]]; then
        SCRIPT_DIR="$(pwd)"
    elif [[ "$SCRIPT_DIR" != /* ]]; then
        SCRIPT_DIR="$(pwd)/$SCRIPT_DIR"
    fi

    # Get the parent directory (project root)
    SOURCE_DIR="$(dirname "$SCRIPT_DIR")"

    # If we can't find the source directory properly, try to find it
    if [[ ! -d "$SOURCE_DIR/backend" || ! -d "$SOURCE_DIR/frontend" ]]; then
        # Try current working directory first
        if [[ -d "$(pwd)/backend" && -d "$(pwd)/frontend" ]]; then
            SOURCE_DIR="$(pwd)"
            log_info "Found MyNodeCP files in current working directory: $SOURCE_DIR"
        # Try the original directory where the script was called from
        elif [[ -n "$OLDPWD" && -d "$OLDPWD/backend" && -d "$OLDPWD/frontend" ]]; then
            SOURCE_DIR="$OLDPWD"
            log_info "Found MyNodeCP files in original directory: $SOURCE_DIR"
        # Try common locations
        elif [[ -d "/root/panelcp/backend" && -d "/root/panelcp/frontend" ]]; then
            SOURCE_DIR="/root/panelcp"
            log_info "Found MyNodeCP files in /root/panelcp: $SOURCE_DIR"
        elif [[ -d "$HOME/panelcp/backend" && -d "$HOME/panelcp/frontend" ]]; then
            SOURCE_DIR="$HOME/panelcp"
            log_info "Found MyNodeCP files in $HOME/panelcp: $SOURCE_DIR"
        else
            SOURCE_DIR="$(pwd)"
            log_warning "Could not find MyNodeCP source files, using current directory: $SOURCE_DIR"
        fi
    fi

    log_info "Script directory: $SCRIPT_DIR"
    log_info "Source directory: $SOURCE_DIR"
    log_info "Current working directory: $(pwd)"

    # Debug: Show what's in the source directory
    log_info "Contents of source directory:"
    ls -la "$SOURCE_DIR" || log_warning "Could not list source directory"

    log_info "Copying files from $SOURCE_DIR to $MYNODECP_HOME..."

    # Copy application files
    if [[ -d "$SOURCE_DIR/backend" && -d "$SOURCE_DIR/frontend" ]]; then
        cp -r "$SOURCE_DIR"/* $MYNODECP_HOME/ 2>/dev/null || {
            log_error "Failed to copy application files"
            log_info "Attempting to copy individual directories..."
            mkdir -p "$MYNODECP_HOME"
            cp -r "$SOURCE_DIR/backend" "$MYNODECP_HOME/" || log_error "Failed to copy backend"
            cp -r "$SOURCE_DIR/frontend" "$MYNODECP_HOME/" || log_error "Failed to copy frontend"
            cp -r "$SOURCE_DIR/scripts" "$MYNODECP_HOME/" 2>/dev/null || log_warning "No scripts directory to copy"
            cp "$SOURCE_DIR"/*.md "$MYNODECP_HOME/" 2>/dev/null || log_warning "No markdown files to copy"
            cp "$SOURCE_DIR"/*.yml "$MYNODECP_HOME/" 2>/dev/null || log_warning "No YAML files to copy"
            cp "$SOURCE_DIR"/*.yaml "$MYNODECP_HOME/" 2>/dev/null || log_warning "No YAML files to copy"
            cp "$SOURCE_DIR"/Makefile "$MYNODECP_HOME/" 2>/dev/null || log_warning "No Makefile to copy"
            cp "$SOURCE_DIR"/Dockerfile "$MYNODECP_HOME/" 2>/dev/null || log_warning "No Dockerfile to copy"
        }
    else
        log_error "MyNodeCP source files not found in $SOURCE_DIR"
        log_info "Please run this script from the MyNodeCP project root directory"
        log_info "Expected structure:"
        log_info "  mynodecp/"
        log_info "  ├── backend/"
        log_info "  ├── frontend/"
        log_info "  └── scripts/install.sh"
        log_info "Current directory contents:"
        ls -la "$SOURCE_DIR"
        exit 1
    fi

    # Verify backend directory exists
    if [[ ! -d "$MYNODECP_HOME/backend" ]]; then
        log_error "Backend directory not found after copying files"
        log_info "Current directory contents:"
        ls -la "$SOURCE_DIR"
        exit 1
    fi

    # Build backend
    log_info "Building backend..."
    cd $MYNODECP_HOME/backend
    export PATH=$PATH:/usr/local/go/bin

    # Initialize Go module if go.mod doesn't exist
    if [[ ! -f "go.mod" ]]; then
        log_info "Initializing Go module..."
        go mod init github.com/mynodecp/mynodecp

        # Add required dependencies
        log_info "Adding Go dependencies..."
        cat > go.mod << 'EOF'
module github.com/mynodecp/mynodecp

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/golang-migrate/migrate/v4 v4.17.0
	github.com/google/uuid v1.5.0
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.0
	github.com/redis/go-redis/v9 v9.3.1
	github.com/spf13/cobra v1.8.0
	github.com/spf13/viper v1.18.2
	github.com/stretchr/testify v1.8.4
	go.uber.org/zap v1.26.0
	golang.org/x/crypto v0.17.0
	google.golang.org/grpc v1.60.1
	google.golang.org/protobuf v1.32.0
	gorm.io/driver/mysql v1.5.2
	gorm.io/gorm v1.25.5
)
EOF
    fi

    go mod download
    go mod tidy
    go build -o mynodecp-server cmd/server/main.go

    # Build frontend
    log_info "Building frontend..."
    cd $MYNODECP_HOME/frontend

    # Check if package.json exists
    if [[ ! -f "package.json" ]]; then
        log_error "Frontend package.json not found"
        exit 1
    fi

    npm install
    npm run build

    # Set permissions
    chown -R $MYNODECP_USER:$MYNODECP_GROUP $MYNODECP_HOME
    chmod +x $MYNODECP_HOME/backend/mynodecp-server

    log_success "MyNodeCP installed"
}

# Create systemd service
create_systemd_service() {
    log_info "Creating systemd service..."

    # Detect Redis service name for systemd dependencies
    if systemctl list-unit-files | grep -q "redis-server.service"; then
        REDIS_SERVICE_NAME="redis-server.service"
    else
        REDIS_SERVICE_NAME="redis.service"
    fi

    cat > /etc/systemd/system/mynodecp.service << EOF
[Unit]
Description=MyNodeCP Control Panel
After=network.target mariadb.service $REDIS_SERVICE_NAME
Wants=mariadb.service $REDIS_SERVICE_NAME

[Service]
Type=simple
User=$MYNODECP_USER
Group=$MYNODECP_GROUP
WorkingDirectory=$MYNODECP_HOME/backend
ExecStart=$MYNODECP_HOME/backend/mynodecp-server
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=mynodecp

Environment=PATH=/usr/local/go/bin:/usr/bin:/bin
Environment=MYNODECP_CONFIG=$MYNODECP_CONFIG

[Install]
WantedBy=multi-user.target
EOF

    # Reload systemd and enable service
    systemctl daemon-reload
    systemctl enable mynodecp

    log_success "Systemd service created"
}

# Configure Nginx reverse proxy
configure_nginx() {
    log_info "Configuring Nginx reverse proxy..."

    cat > /etc/nginx/sites-available/mynodecp << EOF
server {
    listen 80;
    server_name _;

    location / {
        proxy_pass http://localhost:$HTTP_PORT;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF

    # Enable site
    if [[ -d /etc/nginx/sites-enabled ]]; then
        ln -sf /etc/nginx/sites-available/mynodecp /etc/nginx/sites-enabled/
        rm -f /etc/nginx/sites-enabled/default
    else
        # For RHEL-based systems
        cp /etc/nginx/sites-available/mynodecp /etc/nginx/conf.d/mynodecp.conf
    fi

    # Test and reload Nginx
    nginx -t
    systemctl reload nginx

    log_success "Nginx configured"
}

# Main installation function
main() {
    log_info "Starting MyNodeCP installation..."

    check_root
    detect_os
    update_system
    install_dependencies
    install_go
    install_nodejs
    install_mariadb
    install_redis
    install_nginx
    create_user_and_directories
    setup_database
    configure_firewall
    install_mynodecp
    create_systemd_service
    configure_nginx

    # Start MyNodeCP
    systemctl start mynodecp

    # Display installation summary
    display_summary

    log_success "MyNodeCP installation completed!"
}

# Display installation summary
display_summary() {
    echo
    echo "=================================="
    echo "MyNodeCP Installation Summary"
    echo "=================================="
    echo "Version: $MYNODECP_VERSION"
    echo "Installation Directory: $MYNODECP_HOME"
    echo "Configuration Directory: $MYNODECP_CONFIG"
    echo
    echo "Database Information:"
    echo "  Database Name: $DB_NAME"
    echo "  Database User: $DB_USER"
    echo "  Database Password: $DB_PASSWORD"
    echo "  MySQL Root Password: $MYSQL_ROOT_PASSWORD"
    echo
    echo "Access Information:"
    echo "  Control Panel: http://$(hostname -I | awk '{print $1}'):$HTTP_PORT"
    echo "  Direct Access: http://$(hostname -I | awk '{print $1}')"
    echo
    echo "Service Management:"
    echo "  Start:   systemctl start mynodecp"
    echo "  Stop:    systemctl stop mynodecp"
    echo "  Restart: systemctl restart mynodecp"
    echo "  Status:  systemctl status mynodecp"
    echo "  Logs:    journalctl -u mynodecp -f"
    echo
    echo "Configuration Files:"
    echo "  Database: $MYNODECP_CONFIG/database.conf"
    echo "  Main Config: $MYNODECP_HOME/backend/configs/config.yaml"
    echo
    echo "IMPORTANT: Save these passwords in a secure location!"
    echo "=================================="
}

# Run main function
main "$@"
