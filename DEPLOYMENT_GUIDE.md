# MyNodeCP Deployment Guide

## Production Deployment Options

### Option 1: Automated Installation (Recommended)

The fastest way to deploy MyNodeCP in production:

```bash
# Download and run the automated installer
curl -sSL https://raw.githubusercontent.com/mynodecp/mynodecp/main/scripts/install.sh | sudo bash

# Or download first, then run
wget https://raw.githubusercontent.com/mynodecp/mynodecp/main/scripts/install.sh
chmod +x install.sh
sudo ./install.sh
```

This script will:
- Detect your Linux distribution
- Install all required dependencies
- Set up MariaDB and Redis
- Configure Nginx reverse proxy
- Create system user and service
- Configure firewall
- Start MyNodeCP services

### Option 2: Docker Deployment

For containerized deployment:

```bash
# Clone the repository
git clone https://github.com/mynodecp/mynodecp.git
cd mynodecp

# Start the full stack
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f mynodecp
```

### Option 3: Manual Installation

For custom installations or when you need full control:

#### Prerequisites
- Linux server (Ubuntu 20.04+ or CentOS 8+ recommended)
- Minimum 2GB RAM, 10GB disk space
- Root access

#### Step 1: Install Dependencies

**Ubuntu/Debian:**
```bash
apt update
apt install -y curl wget git build-essential

# Install Go 1.21+
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
source /etc/profile

# Install Node.js 18+
curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
apt install -y nodejs

# Install MariaDB
apt install -y mariadb-server mariadb-client

# Install Redis
apt install -y redis-server

# Install Nginx
apt install -y nginx
```

**CentOS/RHEL:**
```bash
yum update -y
yum install -y curl wget git gcc gcc-c++ make

# Install Go 1.21+
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
source /etc/profile

# Install Node.js 18+
curl -fsSL https://rpm.nodesource.com/setup_18.x | bash -
yum install -y nodejs

# Install MariaDB
yum install -y mariadb-server mariadb

# Install Redis
yum install -y redis

# Install Nginx
yum install -y nginx
```

#### Step 2: Setup Database

```bash
# Start MariaDB
systemctl start mariadb
systemctl enable mariadb

# Secure installation
mysql_secure_installation

# Create database and user
mysql -u root -p << EOF
CREATE DATABASE mynodecp;
CREATE USER 'mynodecp'@'localhost' IDENTIFIED BY 'your_secure_password';
GRANT ALL PRIVILEGES ON mynodecp.* TO 'mynodecp'@'localhost';
FLUSH PRIVILEGES;
EOF
```

#### Step 3: Setup Redis

```bash
# Start Redis
systemctl start redis
systemctl enable redis

# Configure Redis (optional)
# Edit /etc/redis/redis.conf for custom settings
```

#### Step 4: Build and Install MyNodeCP

```bash
# Create system user
useradd -r -m -d /opt/mynodecp -s /bin/bash mynodecp

# Clone and build
git clone https://github.com/mynodecp/mynodecp.git /opt/mynodecp/src
cd /opt/mynodecp/src

# Build backend
cd backend
go mod download
go build -o /opt/mynodecp/bin/mynodecp ./cmd/server

# Build frontend
cd ../frontend
npm install
npm run build
cp -r dist/* /opt/mynodecp/public/

# Set permissions
chown -R mynodecp:mynodecp /opt/mynodecp
```

#### Step 5: Configure MyNodeCP

```bash
# Create configuration
mkdir -p /opt/mynodecp/config
cat > /opt/mynodecp/config/config.yaml << EOF
server:
  http_port: 8080
  grpc_port: 9090
  environment: production
  domain: your-domain.com

database:
  host: localhost
  port: 3306
  username: mynodecp
  password: your_secure_password
  database: mynodecp

redis:
  host: localhost
  port: 6379

auth:
  jwt_secret: $(openssl rand -base64 32)
EOF

chown mynodecp:mynodecp /opt/mynodecp/config/config.yaml
chmod 600 /opt/mynodecp/config/config.yaml
```

#### Step 6: Create System Service

```bash
cat > /etc/systemd/system/mynodecp.service << EOF
[Unit]
Description=MyNodeCP Control Panel
After=network.target mariadb.service redis.service

[Service]
Type=simple
User=mynodecp
Group=mynodecp
WorkingDirectory=/opt/mynodecp
ExecStart=/opt/mynodecp/bin/mynodecp
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable mynodecp
systemctl start mynodecp
```

#### Step 7: Configure Nginx

```bash
cat > /etc/nginx/sites-available/mynodecp << EOF
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF

ln -s /etc/nginx/sites-available/mynodecp /etc/nginx/sites-enabled/
rm -f /etc/nginx/sites-enabled/default
nginx -t
systemctl restart nginx
```

## SSL Certificate Setup

### Option 1: Let's Encrypt (Recommended)

```bash
# Install Certbot
apt install -y certbot python3-certbot-nginx  # Ubuntu/Debian
yum install -y certbot python3-certbot-nginx  # CentOS/RHEL

# Get certificate
certbot --nginx -d your-domain.com

# Auto-renewal
echo "0 12 * * * /usr/bin/certbot renew --quiet" | crontab -
```

### Option 2: Custom SSL Certificate

```bash
# Copy your certificates
cp your-certificate.crt /etc/ssl/certs/mynodecp.crt
cp your-private-key.key /etc/ssl/private/mynodecp.key

# Update Nginx configuration
# Add SSL configuration to your Nginx server block
```

## Firewall Configuration

### UFW (Ubuntu/Debian)
```bash
ufw allow 22/tcp
ufw allow 80/tcp
ufw allow 443/tcp
ufw enable
```

### Firewalld (CentOS/RHEL)
```bash
firewall-cmd --permanent --add-service=ssh
firewall-cmd --permanent --add-service=http
firewall-cmd --permanent --add-service=https
firewall-cmd --reload
```

## Post-Installation

### 1. Access MyNodeCP
Open your browser and navigate to:
- HTTP: `http://your-domain.com`
- HTTPS: `https://your-domain.com`

### 2. Initial Setup
- Create admin account
- Configure system settings
- Set up email settings
- Configure backup options

### 3. Service Management

```bash
# Check status
systemctl status mynodecp

# Start/Stop/Restart
systemctl start mynodecp
systemctl stop mynodecp
systemctl restart mynodecp

# View logs
journalctl -u mynodecp -f
```

### 4. Monitoring

Access monitoring dashboards:
- Grafana: `http://your-domain.com:3001` (admin/admin)
- Prometheus: `http://your-domain.com:9091`

## Backup and Recovery

### Database Backup
```bash
# Create backup
mysqldump -u mynodecp -p mynodecp > mynodecp_backup_$(date +%Y%m%d).sql

# Restore backup
mysql -u mynodecp -p mynodecp < mynodecp_backup_20240101.sql
```

### Full System Backup
```bash
# Backup MyNodeCP directory
tar -czf mynodecp_full_backup_$(date +%Y%m%d).tar.gz /opt/mynodecp

# Backup configuration
cp -r /opt/mynodecp/config /backup/mynodecp_config_$(date +%Y%m%d)
```

## Troubleshooting

### Common Issues

1. **Service won't start**
   ```bash
   # Check logs
   journalctl -u mynodecp -n 50
   
   # Check configuration
   /opt/mynodecp/bin/mynodecp --config-check
   ```

2. **Database connection issues**
   ```bash
   # Test database connection
   mysql -u mynodecp -p -h localhost mynodecp
   
   # Check MariaDB status
   systemctl status mariadb
   ```

3. **Permission issues**
   ```bash
   # Fix permissions
   chown -R mynodecp:mynodecp /opt/mynodecp
   chmod 755 /opt/mynodecp/bin/mynodecp
   ```

### Performance Tuning

1. **Database Optimization**
   ```bash
   # Edit /etc/mysql/mariadb.conf.d/50-server.cnf
   innodb_buffer_pool_size = 256M
   max_connections = 200
   query_cache_size = 64M
   ```

2. **Redis Optimization**
   ```bash
   # Edit /etc/redis/redis.conf
   maxmemory 256mb
   maxmemory-policy allkeys-lru
   ```

## Support and Updates

### Getting Help
- Documentation: https://docs.mynodecp.com
- Community Forum: https://community.mynodecp.com
- GitHub Issues: https://github.com/mynodecp/mynodecp/issues

### Updates
```bash
# Check for updates
curl -s https://api.github.com/repos/mynodecp/mynodecp/releases/latest

# Update MyNodeCP
./scripts/update.sh
```

### Commercial Support
For enterprise support and custom installations:
- Email: support@mynodecp.com
- Website: https://mynodecp.com/support
