# MyNodeCP Installation Fixes

## Issues Resolved

### 1. Redis Service Name Issue ✅
**Problem**: The installation script was trying to enable `redis.service` but on some systems it's named `redis-server.service`.

**Solution**: 
- Added dynamic Redis service detection
- Script now checks for both `redis.service` and `redis-server.service`
- Falls back gracefully if service detection fails
- Updated systemd service dependencies to use the correct Redis service name

### 2. MySQL Interactive Prompts ✅
**Problem**: MariaDB installation was prompting for user input during installation.

**Solution**:
- Added `DEBIAN_FRONTEND=noninteractive` environment variable
- Pre-configured MariaDB root password using `debconf-set-selections`
- Automated MySQL secure installation process
- Non-interactive password setting for root user
- Automatic removal of test databases and anonymous users

### 3. Enhanced Error Handling ✅
**Additional Improvements**:
- Better error handling with fallback options
- Comprehensive logging with color-coded output
- Automatic password generation for security
- Installation summary with all credentials displayed
- Service status verification

## Updated Installation Script Features

### Non-Interactive Installation
```bash
# The script now runs completely non-interactively
sudo ./scripts/install.sh
```

### Automatic Password Generation
- MySQL root password: Automatically generated (32-character base64)
- MyNodeCP database password: Automatically generated (32-character base64)
- All passwords displayed at the end of installation

### Service Detection
- Automatically detects Redis service name (`redis` vs `redis-server`)
- Handles different Linux distributions properly
- Graceful fallback for service management

### Installation Summary
At the end of installation, you'll see:
```
==================================
MyNodeCP Installation Summary
==================================
Version: 1.0.0
Installation Directory: /opt/mynodecp
Configuration Directory: /etc/mynodecp

Database Information:
  Database Name: mynodecp
  Database User: mynodecp
  Database Password: [generated-password]
  MySQL Root Password: [generated-password]

Access Information:
  Control Panel: http://[your-ip]:8080
  Direct Access: http://[your-ip]

Service Management:
  Start:   systemctl start mynodecp
  Stop:    systemctl stop mynodecp
  Restart: systemctl restart mynodecp
  Status:  systemctl status mynodecp
  Logs:    journalctl -u mynodecp -f

IMPORTANT: Save these passwords in a secure location!
==================================
```

## Additional Scripts Created

### 1. Troubleshooting Script
```bash
# Run comprehensive diagnostics
sudo ./scripts/troubleshoot.sh

# Check specific components
sudo ./scripts/troubleshoot.sh services
sudo ./scripts/troubleshoot.sh database
sudo ./scripts/troubleshoot.sh redis
sudo ./scripts/troubleshoot.sh permissions

# Fix common issues
sudo ./scripts/troubleshoot.sh fix
sudo ./scripts/troubleshoot.sh restart
```

### 2. Update Script
```bash
# Update to latest version
sudo ./scripts/update.sh

# Check for updates
sudo ./scripts/update.sh check

# Rollback if needed
sudo ./scripts/update.sh rollback
```

## Testing the Fixed Installation

### Prerequisites
- Fresh Linux server (Ubuntu 20.04+, CentOS 8+, etc.)
- Root access
- Internet connection

### Installation Steps
1. **Download MyNodeCP**:
   ```bash
   git clone https://github.com/mynodecp/mynodecp.git
   cd mynodecp
   ```

2. **Make scripts executable**:
   ```bash
   chmod +x scripts/*.sh
   ```

3. **Run installation**:
   ```bash
   sudo ./scripts/install.sh
   ```

4. **Verify installation**:
   ```bash
   sudo ./scripts/troubleshoot.sh
   ```

### Expected Results
- ✅ No interactive prompts during installation
- ✅ All services start correctly
- ✅ Redis service detected and started properly
- ✅ MySQL configured with secure passwords
- ✅ MyNodeCP accessible via web browser
- ✅ All credentials displayed at end of installation

## Common Issues and Solutions

### Issue: Redis service fails to start
**Solution**: 
```bash
# Check which Redis service is available
systemctl list-unit-files | grep redis

# Start the correct service manually
sudo systemctl start redis-server  # or redis
sudo systemctl enable redis-server  # or redis
```

### Issue: MySQL connection fails
**Solution**:
```bash
# Check MySQL status
sudo systemctl status mariadb

# Test connection with generated credentials
mysql -u mynodecp -p[password] mynodecp

# Reset password if needed
sudo mysql -u root -p[root-password]
```

### Issue: MyNodeCP service won't start
**Solution**:
```bash
# Check service logs
sudo journalctl -u mynodecp -f

# Verify binary permissions
sudo chmod +x /opt/mynodecp/backend/mynodecp-server

# Fix ownership
sudo chown -R mynodecp:mynodecp /opt/mynodecp
```

## Security Improvements

1. **Strong Password Generation**: All passwords are 32-character base64 encoded
2. **Secure File Permissions**: Configuration files have 600 permissions
3. **Service User**: MyNodeCP runs as dedicated `mynodecp` user
4. **Firewall Configuration**: Only necessary ports are opened
5. **Database Security**: Test databases and anonymous users removed

## Next Steps

After successful installation:

1. **Access the control panel** at `http://your-server-ip`
2. **Complete initial setup** with admin account creation
3. **Configure SSL certificate** using Let's Encrypt
4. **Set up monitoring** with the included Grafana/Prometheus stack
5. **Create your first hosting account** and domain

## Support

If you encounter any issues:

1. **Run diagnostics**: `sudo ./scripts/troubleshoot.sh`
2. **Check logs**: `sudo journalctl -u mynodecp -f`
3. **Review configuration**: Check `/etc/mynodecp/database.conf`
4. **Get help**: Visit https://community.mynodecp.com

The installation script is now production-ready and handles all the common issues encountered during deployment! 🚀
